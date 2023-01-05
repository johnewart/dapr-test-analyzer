package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"test-analyzer/analysis"
	"test-analyzer/ingestion"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := ingestion.NewRecordDB()

	if db == nil {
		fmt.Println("Unable to open DB...")
	} else {
		groups := db.AllReportGroups()
		for i, rg := range groups {
			fmt.Printf("Processing report group %d of %d\r", i+1, len(db.AllReportGroups()))
			reportIdsWithData := make(map[uint]bool)
			for _, id := range db.GetReportIDsWithTestMetrics() {
				reportIdsWithData[id] = true
			}

			fullGroup := db.LoadReportGroup(rg.ID)
			for _, report := range fullGroup.Reports {
				if _, exists := reportIdsWithData[report.ID]; !exists {
					fmt.Printf("Generating test metrics for report %d\n", report.ID)
					result := analysis.GenerateTestMetrics(report)
					for _, tm := range result {
						rtm := ingestion.ReportTestMetrics{
							ReportID:  report.ID,
							TestLabel: tm.TestLabel,
							PassCount: tm.PassCount,
							FailCount: tm.FailCount,
						}

						db.StoreReportTestMetrics(rtm)
					}
				}
			}
		}
	}
}
