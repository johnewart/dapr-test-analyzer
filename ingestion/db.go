package ingestion

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type RecordDB struct {
	Path string
	db   *gorm.DB
}

func NewRecordDB() *RecordDB {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)

	if db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Logger: newLogger,
	}); err != nil {
		panic("failed to connect database")
	} else {
		if err := db.AutoMigrate(&Project{}, &ReportGroup{}, &Report{}, &TestGroup{}, &Test{}, &TestLog{}, &ReportTestMetrics{}); err != nil {
			fmt.Printf("Error migrating: %v\n", err)
			return nil
		} else {
			return &RecordDB{
				Path: "gorm.db",
				db:   db,
			}
		}
	}
}

func (r *RecordDB) FindProjectByLabel(label string) *Project {
	var project Project
	if tx := r.db.First(&project, "label = ?", label); tx.Error != nil {
		return nil
	} else {
		return &project
	}
}

func (r *RecordDB) FindReportGroupByLabel(projectId uint, label string) *ReportGroup {
	var reportGroup ReportGroup
	if tx := r.db.First(&reportGroup, "project_id = ? AND label = ?", projectId, label); tx.Error != nil {
		return nil
	} else {
		return &reportGroup
	}
}

func (r *RecordDB) FindReportByLabel(reportGroupId uint, label string) *Report {
	var report Report
	if tx := r.db.First(&report, "report_group_id = ? AND label = ?", reportGroupId, label); tx.Error != nil {
		return nil
	} else {
		return &report
	}
}

func (r *RecordDB) FindTestGroupByLabel(reportId uint, label string) *TestGroup {
	var testGroup TestGroup
	if tx := r.db.First(&testGroup, "report_id = ? AND label = ?", reportId, label); tx.Error != nil {
		return nil
	} else {
		return &testGroup
	}
}

func (r *RecordDB) FindTestByLabel(testGroupId uint, label string) *Test {
	var test Test
	if tx := r.db.First(&test, "test_group_id = ? AND label = ?", testGroupId, label); tx.Error != nil {
		return nil
	} else {
		return &test
	}
}

func (r *RecordDB) StoreReportGroup(reportGroup ReportGroup) (uint, error) {
	tx := r.db.Create(&reportGroup).Clauses(clause.OnConflict{DoNothing: true})
	return reportGroup.ID, tx.Error
}

func (r *RecordDB) StoreReport(report Report) (uint, error) {
	tx := r.db.Create(&report).Clauses(clause.OnConflict{DoNothing: true})
	return report.ID, tx.Error
}

func (r *RecordDB) StoreTestGroup(testGroup TestGroup) (uint, error) {
	tx := r.db.Create(&testGroup).Clauses(clause.OnConflict{DoNothing: true})
	return testGroup.ID, tx.Error
}

func (r *RecordDB) StoreTest(test Test) (uint, error) {
	tx := r.db.Create(&test).Clauses(clause.OnConflict{DoNothing: true})
	return test.ID, tx.Error
}

func (r *RecordDB) StoreTestLog(log TestLog) {
	r.db.Create(&log).Clauses(clause.OnConflict{DoNothing: true})
}

func (r *RecordDB) AllReportGroups() []ReportGroup {
	var reportGroups []ReportGroup
	r.db.Find(&reportGroups)
	return reportGroups
}

func (r *RecordDB) GetReportGroup() ReportGroup {
	var reportGroup ReportGroup
	r.db.First(&reportGroup)
	return reportGroup
}

func (r *RecordDB) GetReports(reportGroupId uint) []Report {
	var reports []Report
	r.db.Where("report_group_id = ?", reportGroupId).Find(&reports)
	return reports
}

func (r *RecordDB) GetTestGroups(reportId uint) []TestGroup {
	var testGroups []TestGroup
	r.db.Where("report_id = ?", reportId).Find(&testGroups)
	return testGroups
}

func (r *RecordDB) GetTests(testGroupId uint) []Test {
	var tests []Test
	r.db.Where("test_group_id = ?", testGroupId).Find(&tests)
	return tests
}

func (r *RecordDB) GetTestLogs(testId uint) []TestLog {
	var logs []TestLog
	r.db.Where("test_id = ?", testId).Find(&logs)
	return logs
}

func (r *RecordDB) LoadReportGroup(reportGroupId uint) *ReportGroup {
	var reportGroup ReportGroup
	r.db.Where("id = ?", reportGroupId).First(&reportGroup)
	reportGroup.Reports = r.GetReports(reportGroupId)
	for l, report := range reportGroup.Reports {
		report.TestGroups = r.GetTestGroups(report.ID)
		for i, testGroup := range report.TestGroups {
			report.TestGroups[i].Tests = r.GetTests(testGroup.ID)
			for j, test := range report.TestGroups[i].Tests {
				report.TestGroups[i].Tests[j].Logs = r.GetTestLogs(test.ID)
			}
		}
		reportGroup.Reports[l] = report
	}
	return &reportGroup
}

func (r *RecordDB) StoreReportTestMetrics(metrics ReportTestMetrics) {
	r.db.Create(&metrics).Clauses(clause.OnConflict{DoNothing: true})
}

func (r *RecordDB) GetReportTestMetrics(reportId uint) []ReportTestMetrics {
	var metrics []ReportTestMetrics
	r.db.Where("report_id = ?", reportId).Find(&metrics)
	return metrics
}

func (r *RecordDB) GetReportIDsWithTestMetrics() []uint {
	var ids []uint
	r.db.Model(&ReportTestMetrics{}).Distinct("report_id").Pluck("report_id", &ids)
	return ids
}

func (r *RecordDB) GetReportIDsWithoutTestMetrics() []uint {
	var ids []uint
	r.db.Model(&Report{}).Where("id NOT IN (?)", r.GetReportIDsWithTestMetrics()).Pluck("id", &ids)
	return ids
}

func (r *RecordDB) FindOrCreateReportGroupByLabel(projectId uint, label string) *ReportGroup {
	if reportGroup := r.FindReportGroupByLabel(projectId, label); reportGroup != nil {
		return reportGroup
	} else {
		rg := ReportGroup{
			ProjectID: projectId,
			Label:     label,
		}
		if id, err := r.StoreReportGroup(rg); err != nil {
			fmt.Printf("Error creating report group: %v\n", err)
			return nil
		} else {
			rg.ID = id
			return &rg
		}
	}
}
