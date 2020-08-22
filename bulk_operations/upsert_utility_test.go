package upsertutility_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type TestObj struct {
	ID        int32
	Name      string
	UpdatedBy string
	CreatedBy string
}

var errFailed = errors.New("failed")

// GetDB provides mock DB objects for testing
func GetDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mocked db can't be error")
	}

	gormDB, err := gorm.Open("mysql", db)
	if err != nil {
		t.Fatalf("error creating gorm DB")
	}

	f := func() {
		_ = gormDB.Close()
		_ = db.Close()
	}
	return gormDB, mock, f
}

func TestBatchUpsert_Records(t *testing.T) {
	dataArr := []TestObj{
		{
			ID:   1,
			Name: "name",
		},
		{
			ID:   2,
			Name: "name2",
		},
	}

	// getting mock db instance
	gormDB, mock, deferFn := GetDB(t)
	gormDB = gormDB.Set("audit:current_user", "user@test.com")

	var dataInterfaceArr []interface{}
	for _, data := range dataArr {
		dataInterfaceArr = append(dataInterfaceArr, data)
	}

	mock.ExpectExec("INSERT INTO  `test_obj`").
		WithArgs(
			dataArr[0].ID,
			dataArr[0].Name,
			dataArr[1].ID,
			dataArr[1].Name,
		).
		WillReturnResult(sqlmock.NewResult(1, 1)).
		WillReturnError(nil)
	bulkUpsertStats, err := upsertutiltiy.BatchUpsert(gormDB, dataInterfaceArr, false)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, bulkUpsertStats.UpsertFailCount)
	assert.Equal(t, 2, bulkUpsertStats.UpsertCount)
	deferFn()
}

func TestBatchUpsert_SingleRecordError(t *testing.T) {
	dataArr := []TestObj{
		{
			ID:   1,
			Name: "name",
		},
	}

	// getting mock db instance
	gormDB, mock, deferFn := GetDB(t)
	gormDB = gormDB.Set("audit:current_user", "user@test.com")

	var dataInterfaceArr []interface{}
	for _, data := range dataArr {
		dataInterfaceArr = append(dataInterfaceArr, data)
	}

	mock.ExpectExec("INSERT IGNORE INTO  `test_obj`").
		WithArgs(
			dataArr[0].ID,
			dataArr[0].Name,
		).
		WillReturnResult(nil).
		WillReturnError(errFailed)
	bulkUpsertStats, err := upsertutiltiy.BatchUpsert(gormDB, dataInterfaceArr, true)
	assert.Equal(t, errFailed, err)
	assert.Equal(t, 1, bulkUpsertStats.UpsertFailCount)
	assert.Equal(t, 0, bulkUpsertStats.UpsertCount)
	deferFn()
}

func TestBatchUpsert_NoRecord(t *testing.T) {
	dataArr := []TestObj{}

	// getting mock db instance
	gormDB, _, deferFn := GetDB(t)

	var dataInterfaceArr []interface{}
	for _, data := range dataArr {
		dataInterfaceArr = append(dataInterfaceArr, data)
	}

	bulkUpsertStats, err := upsertutiltiy.BatchUpsert(gormDB, dataInterfaceArr, false)
	assert.Nil(t, err)
	assert.Nil(t, bulkUpsertStats)
	deferFn()
}
