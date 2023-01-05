package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"test-analyzer/analysis"
	"test-analyzer/ingestion"
)

var db *ingestion.RecordDB

type ReportResponse struct {
	Report *ingestion.Report `json:"report"`
}

func toHTML(s string) template.HTML {

	// Escape everything in the string first to ensure that
	// special characters ('<' for example) are displayed as
	// characters and not treated as markup.
	s = template.HTMLEscapeString(s)

	// Insert the links.
	//s = linkPat.ReplaceAllStringFunc(s, func(m string) string {
	//	s = s[1 : len(s)-1]
	//	return "<a href='/view/" + m + "'>" + m + "</a>"
	//})

	return template.HTML(s)
}

type Page struct {
	Title string
	Body  string
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles("templates/" + tmpl + ".html")
	_ = t.Execute(w, p)
}

func GetReport(w http.ResponseWriter, r *http.Request) {
	//id := mux.Vars(r)["id"]

	report := db.GetReportGroup()
	fullReport := db.LoadReportGroup(report.ID)

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(fullReport)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadOrGenerateMetrics() []analysis.TestMetrics {
	var testMetrics []analysis.TestMetrics
	fname := "report.json"
	fileExists := true
	rebuild := false

	if _, err := os.Stat("report.json"); err != nil {
		if os.IsNotExist(err) {
			fileExists = false
		}
	}

	if fileExists {
		if f, err := os.Open(fname); err != nil {
			fmt.Printf("Unable to open report.json: %v\n", err)
			rebuild = true
		} else {
			if b, err := io.ReadAll(f); err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				rebuild = true
			} else {
				if len(b) == 0 {
					fmt.Println("Empty file...")
					rebuild = true
				} else {
					fmt.Printf("Read %d bytes\n", len(b))
					if err := json.Unmarshal(b, &testMetrics); err != nil {
						fmt.Printf("Failed to parse JSON: %v\n", err)
						rebuild = true
					}
				}
			}
			f.Close()
		}

		if rebuild == true {
			fmt.Println("Removing report file")
			os.Remove(fname)
		}
	}

	if rebuild {
		testMetrics = make([]analysis.TestMetrics, 0)
		fmt.Printf("Rebuilding report data...\n")
		reportIds := db.GetReportIDsWithTestMetrics()
		testMetricsMap := make(map[string][]analysis.TestMetrics, len(reportIds))

		for i, rid := range reportIds {
			fmt.Printf("[%d/%d] Processing report %d\r", i, len(reportIds), rid)
			for _, tm := range db.GetReportTestMetrics(rid) {
				if _, ok := testMetricsMap[tm.TestLabel]; !ok {
					testMetricsMap[tm.TestLabel] = make([]analysis.TestMetrics, 0)
				}
				testMetricsMap[tm.TestLabel] = append(testMetricsMap[tm.TestLabel], analysis.TestMetrics{
					TestLabel: tm.TestLabel,
					PassCount: tm.PassCount,
					FailCount: tm.FailCount,
				})
			}
		}

		for _, metrics := range testMetricsMap {
			tm := analysis.TestMetrics{
				TestLabel: metrics[0].TestLabel,
				PassCount: 0,
				FailCount: 0,
			}
			for _, m := range metrics {
				tm = tm.Add(m)
			}
			testMetrics = append(testMetrics, tm)
		}

		if of, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0600); err != nil {
			fmt.Printf("Failed to open %s for writing: %v\n", fname, err)
		} else {
			if b, err := json.Marshal(testMetrics); err != nil {
				fmt.Printf("Failed to marshal metrics to json: %v\n", err)
			} else {
				fmt.Printf("Writing data to %s\n", fname)
				of.Write(b)
			}
			of.Close()
		}
	}

	tm := make([]analysis.TestMetrics, 0)
	for _, r := range testMetrics {
		r.TestLabel = strings.ReplaceAll(r.TestLabel, ">", "&gt;")
		r.TestLabel = strings.ReplaceAll(r.TestLabel, "<", "&lt;")
		tm = append(tm, r)
	}

	return tm
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	testMetrics := loadOrGenerateMetrics()
	if b, err := json.Marshal(testMetrics); err != nil {
		w.WriteHeader(500)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

func GetIndex(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/test_report.html")
	testMetrics := loadOrGenerateMetrics()

	sort.Slice(testMetrics, func(i int, j int) bool {
		return testMetrics[j].PassRate() > testMetrics[i].PassRate()
	})

	fmt.Printf("Rendering template with %d test metrics\n", len(testMetrics))
	data := struct {
		TestMetrics []analysis.TestMetrics
	}{
		TestMetrics: testMetrics,
	}

	_ = t.Execute(w, &data)
}

func loadTestHistory() []analysis.TestHistory {
	fname := "heatmap.json"
	var result []analysis.TestHistory

	if _, err := os.Stat(fname); os.IsNotExist(err) {
		result = make([]analysis.TestHistory, 0)
		for _, g := range db.AllReportGroups() {
			fmt.Printf("Loading report %d...", g.ID)
			rg := db.LoadReportGroup(g.ID)
			reportData := analysis.GenerateTestHistory(rg)
			for _, h := range reportData {
				result = append(result, h)
			}
		}
		fmt.Println()

		if f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0600); err != nil {
			fmt.Printf("Failed to create file %s: %v\n", fname, err)
		} else {
			if b, err := json.Marshal(result); err != nil {
				fmt.Printf("Failed to marshal: %v\n", err)
			} else {
				f.Write(b)
			}
			f.Close()
		}
	} else {
		if f, err := os.Open(fname); err == nil {
			if b, err := io.ReadAll(f); err == nil {
				json.Unmarshal(b, &result)
			}
		}
	}

	return result
}

func GetHeatmapData(w http.ResponseWriter, r *http.Request) {

	if b, err := json.Marshal(loadTestHistory()); err != nil {
		w.WriteHeader(500)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

func GetHeatmap(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/heatmap.html")
	testMetrics := loadOrGenerateMetrics()

	sort.Slice(testMetrics, func(i int, j int) bool {
		return testMetrics[j].PassRate() > testMetrics[i].PassRate()
	})

	fmt.Printf("Rendering template with %d test metrics\n", len(testMetrics))
	data := struct {
		TestMetrics []analysis.TestMetrics
	}{
		TestMetrics: testMetrics,
	}

	_ = t.Execute(w, &data)
}

func ServeHTTP() {
	db = ingestion.NewRecordDB()
	if db == nil {
		log.Fatal("Unable to open DB...")
	}

	r := mux.NewRouter()

	r.HandleFunc("/table", GetIndex).Methods("GET")
	r.HandleFunc("/report", GetReport).Methods("GET")
	r.HandleFunc("/", GetHeatmap).Methods("GET")
	r.HandleFunc("/heatmap", GetHeatmap).Methods("GET")
	r.HandleFunc("/heatmap.json", GetHeatmapData).Methods("GET")
	r.HandleFunc("/report.json", GetMetrics).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	fmt.Println("Listening and serving on http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", r))
}
