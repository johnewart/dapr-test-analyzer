package analysis

import (
	"test-analyzer/ingestion"
)

type ReportMetrics struct {
	ReportID    uint
	ReportLabel string
	PassCount   uint
	FailCount   uint
}

type TestMetrics struct {
	TestLabel string
	PassCount uint
	FailCount uint
}

func (t TestMetrics) PassRate() float64 {
	return float64(t.PassCount) / float64(t.PassCount+t.FailCount)
}

func (t TestMetrics) Add(o TestMetrics) TestMetrics {
	t.PassCount += o.PassCount
	t.FailCount += o.FailCount
	return t
}

func (m ReportMetrics) PassRate() float64 {
	return float64(m.PassCount) / float64(m.PassCount+m.FailCount)
}

func GenerateTestMetrics(report ingestion.Report) []TestMetrics {
	dataMap := make(map[string]TestMetrics)
	for _, group := range report.TestGroups {
		for _, test := range group.Tests {
			if _, ok := dataMap[test.Label]; !ok {
				dataMap[test.Label] = TestMetrics{
					TestLabel: test.Label,
					PassCount: 0,
					FailCount: 0,
				}
			}

			if test.Status == "pass" {
				dataMap[test.Label] = TestMetrics{
					TestLabel: test.Label,
					PassCount: dataMap[test.Label].PassCount + 1,
					FailCount: dataMap[test.Label].FailCount,
				}
			} else if test.Status == "fail" {
				dataMap[test.Label] = TestMetrics{
					TestLabel: test.Label,
					PassCount: dataMap[test.Label].PassCount,
					FailCount: dataMap[test.Label].FailCount + 1,
				}
			}
		}
	}

	results := make([]TestMetrics, 0, len(dataMap))
	for _, v := range dataMap {
		results = append(results, v)
	}
	return results
}

func GenerateMetrics(reportGroup ingestion.ReportGroup) map[uint]ReportMetrics {
	result := make(map[uint]ReportMetrics)
	for _, report := range reportGroup.Reports {

		reportPassCount := 0
		reportFailCount := 0
		for _, group := range report.TestGroups {
			passCount := 0
			failCount := 0
			for _, test := range group.Tests {
				if test.Status == "pass" {
					passCount++
				} else if test.Status == "fail" {
					failCount++
				}
			}

			reportPassCount += passCount
			reportFailCount += failCount
		}
		result[report.ID] = ReportMetrics{
			ReportID:    report.ID,
			ReportLabel: report.Label,
			PassCount:   uint(reportPassCount),
			FailCount:   uint(reportFailCount),
		}
	}

	return result
}

type TestHistory struct {
	Label    string
	Group    string
	Subgroup string
	Passed   bool
}

func GenerateTestHistory(reportGroup *ingestion.ReportGroup) []TestHistory {
	result := make([]TestHistory, 0)
	for _, report := range reportGroup.Reports {
		for _, group := range report.TestGroups {
			for _, test := range group.Tests {
				result = append(result, TestHistory{
					Label:    test.Label,
					Group:    reportGroup.Label,
					Subgroup: report.Label,
					Passed:   test.Status == "pass",
				})
			}
		}
	}

	return result
}
