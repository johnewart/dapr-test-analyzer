package ingestion

import (
	"gorm.io/gorm"
	"math"
	"time"
)

type Project struct {
	gorm.Model

	Label        string `gorm:"uniqueIndex"`
	ReportGroups []ReportGroup
}

type ReportGroup struct {
	gorm.Model

	ProjectID uint   `gorm:"index:idx_project_id_label"`
	Label     string `gorm:"index:idx_project_id_label"`
	Reports   []Report
}

func (rg ReportGroup) TimeWindow() TimeWindow {
	var min int64 = math.MaxInt64
	var max int64 = 0

	for _, r := range rg.Reports {
		tw := r.TimeWindow()
		if tw.Start < min {
			min = tw.Start
		}
		if tw.End > max {
			max = tw.End
		}
	}

	return TimeWindow{
		Start: min,
		End:   max,
	}
}

type Report struct {
	gorm.Model

	ReportGroupID uint   `gorm:"index:idx_report_group_id_label"`
	Label         string `gorm:"index:idx_report_group_id_label"`
	TestGroups    []TestGroup
}

func (r Report) TimeWindow() TimeWindow {
	var min int64 = math.MaxInt64
	var max int64 = 0

	for _, tg := range r.TestGroups {
		tw := tg.TimeWindow()
		if tw.Start < min {
			min = tw.Start
		}
		if tw.End > max {
			max = tw.End
		}
	}

	return TimeWindow{
		Start: min,
		End:   max,
	}
}

type ReportTestMetrics struct {
	gorm.Model

	ReportID  uint
	TestLabel string
	PassCount uint
	FailCount uint
}

func (r ReportTestMetrics) PassRate() float64 {
	return float64(r.PassCount) / float64(r.PassCount+r.FailCount)
}

func (r ReportTestMetrics) FailRate() float64 {
	return 1 - r.PassRate()
}

type TestGroup struct {
	gorm.Model

	Label    string
	ReportID uint `gorm:"index:idx_testgroup_report_id"`
	Tests    []Test
}

func (tg TestGroup) TimeWindow() TimeWindow {
	var min int64 = math.MaxInt64
	var max int64 = 0

	for _, t := range tg.Tests {
		if t.Start < min {
			min = t.Start
		}
		if t.End > max {
			max = t.End
		}
	}

	return TimeWindow{
		Start: min,
		End:   max,
	}
}

type TimeWindow struct {
	Start int64
	End   int64
}

type Test struct {
	gorm.Model

	TestGroupID uint `gorm:"index:idx_test_test_group_id"`
	Label       string
	Status      string
	Start       int64
	End         int64
	Logs        []TestLog
}

type TestLog struct {
	gorm.Model

	TestID    uint  `gorm:"index:idx_test_timestamp"`
	Timestamp int64 `gorm:"index:idx_test_timestamp"`
	Text      string
}

type TestRecord struct {
	Time    string  `json:"time,omitempty"`
	Action  string  `json:"action,omitempty"`
	Package string  `json:"package,omitempty"`
	Output  string  `json:"output,omitempty"`
	Elapsed float64 `json:"elapsed,omitempty"`
	Test    string  `json:"test,omitempty"`
}

func (t TestRecord) Timestamp() int64 {
	if tt, err := time.Parse(time.RFC3339, t.Time); err != nil {
		return -1
	} else {
		return tt.UnixMilli()
	}
}
