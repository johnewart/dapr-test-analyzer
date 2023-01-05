package ingestion

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func extractReport(zipfilePath string) ([]byte, error) {
	if r, err := zip.OpenReader(zipfilePath); err != nil {
		return nil, err
	} else {
		defer r.Close()

		for _, f := range r.File {
			if strings.Contains(f.Name, "e2e") {
				fmt.Printf("Found report file %s\n", f.Name)
				if rc, err := f.Open(); err != nil {
					return nil, err
				} else {
					defer rc.Close()

					return io.ReadAll(rc)
				}
			}
		}

		return nil, nil
	}
}

func GenerateReport(reportLabel string, db *RecordDB, records []TestRecord) Report {
	report := Report{
		Label: reportLabel,
	}
	groupTestMap := make(map[string]map[string]Test)
	for _, rec := range records {

		if rec.Package != "" {
			if _, ok := groupTestMap[rec.Package]; !ok {
				groupTestMap[rec.Package] = make(map[string]Test)
			}
			if rec.Test != "" {
				if t, ok := groupTestMap[rec.Package][rec.Test]; !ok {
					t = Test{
						Label: rec.Test,
					}
					groupTestMap[rec.Package][rec.Test] = t
				}

				if t, ok := groupTestMap[rec.Package][rec.Test]; ok {
					if t.Start == 0 || t.Start > rec.Timestamp() {
						t.Start = rec.Timestamp()
					}

					if t.End == 0 || t.End < rec.Timestamp() {
						t.End = rec.Timestamp()
					}

					if rec.Action == "fail" || rec.Action == "skip" || rec.Action == "pass" {
						t.Status = rec.Action
					}

					if rec.Action == "output" {
						t.Logs = append(t.Logs, TestLog{
							Timestamp: rec.Timestamp(),
							Text:      rec.Output,
						})
					}

					groupTestMap[rec.Package][rec.Test] = t
				}
			}
		}
	}

	for groupLabel, group := range groupTestMap {
		testGroup := TestGroup{
			Label: groupLabel,
		}
		for _, test := range group {
			testGroup.Tests = append(testGroup.Tests, test)
		}
		report.TestGroups = append(report.TestGroups, testGroup)
	}

	return report
}

func ExtractResults(db *RecordDB) error {
	reportPath := os.Getenv("CACHE_DIR")
	if reportPath == "" {
		if userDir, err := os.UserHomeDir(); err != nil {
			return fmt.Errorf("unable to determine cache dir via environment or default ~/.cache/dapr-test-analyzer")
		} else {
			reportPath = path.Join(userDir, ".cache", "dapr-test-analyzer")
		}
	}

	store := ArtifactStore{RootPath: reportPath}
	projectId := uint(1)

	if artifacts, err := store.ListArtifacts(); err != nil {
		return fmt.Errorf("failed to list artifacts: %w", err)
	} else {
		fmt.Printf("Found %d artifacts\n", len(artifacts))

		for _, artifact := range artifacts {
			fmt.Printf("Extracting %v\n", *artifact.Name)
			if b, err := extractReport(store.PathToArtifact(artifact)); err != nil {
				return fmt.Errorf("failed to extract report: %w", err)
			} else {
				fmt.Printf("Extracted %d bytes from %s\n", len(b), *artifact.Name)
				fData := string(b[:])
				lines := strings.Split(fData, "\n")
				errorCount := 0
				records := make([]TestRecord, 0)
				for _, line := range lines {
					var rec TestRecord
					if err := json.Unmarshal([]byte(line), &rec); err != nil {
						errorCount += 1
					} else {
						records = append(records, rec)
					}
				}
				fmt.Printf("Extracted %d records from %s with %d errors\n", len(records), *artifact.Name, errorCount)
				runId := fmt.Sprintf("%d", *artifact.WorkflowRunMetadata.ID)
				report := GenerateReport(*artifact.Name, db, records)
				groupLabel := "Workflow Run " + runId

				reportGroup := db.FindOrCreateReportGroupByLabel(projectId, groupLabel)

				report.ReportGroupID = reportGroup.ID

				//fmt.Printf("Report: %v\n", report)
				if reportId, err := db.StoreReport(report); err == nil {
					fmt.Printf("Stored report %d\n", reportId)
					/*for _, g := range report.TestGroups {
						fmt.Printf("TestGroup: %d\n", g.ReportID)
						g.ReportID = reportId
						fmt.Printf("TestGroup: %d\n", g.ReportID)
						if groupId, err := db.StoreTestGroup(g); err == nil {
							fmt.Printf("Stored test group %d\n", groupId)
							for _, t := range g.Tests {
								t.TestGroupID = groupId
								if testId, err := db.StoreTest(t); err == nil {
									for _, l := range t.Logs {
										l.TestID = testId
										db.StoreTestLog(l)
									}
								}
							}
						}
					}*/
				}
			}
		}
	}

	return nil
}
