package audit

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

type product struct {
	Name string

	Model
}

type productWithoutAuditFields struct {
	Name string
}

func TestCallbacks_Save(t *testing.T) {

	t.Run("save", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`,`created_by`,`updated_by`) VALUES (?,?,?)")).
			WithArgs("product1", "user@test.com", "user@test.com").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Save(&product{Name: "product1"}).Error)
	})

	t.Run("save no user", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`,`created_by`,`updated_by`) VALUES (?,?,?)")).
			WithArgs("product1", "", "").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB.Save(&product{Name: "product1"})
		assert.NotNil(t, gormDB.Save(&product{Name: "product1"}).Error)
	})

	t.Run("save with productWithoutAuditFields", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`) VALUES (?)")).
			WithArgs("product1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Save(&productWithoutAuditFields{Name: "product1"}).Error)
	})

	t.Run("save nil object", func(t *testing.T) {
		gormDB, _, defeFunc := getDB(t)
		defer defeFunc()
		assert.NotNil(t, gormDB.Save(nil))
	})
	t.Run("create", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`,`created_by`,`updated_by`) VALUES (?,?,?)")).
			WithArgs("product1", "user@test.com", "user@test.com").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Create(&product{Name: "product1"}))
	})
	t.Run("create with no user", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`,`created_by`,`updated_by`) VALUES (?,?,?)")).
			WithArgs("product1", "", "").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		assert.NotNil(t, gormDB.Create(&product{Name: "product1"}))
	})
	t.Run("create with productWithoutAuditFields", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("INSERT  INTO `products` (`name`) VALUES (?)")).
			WithArgs("product1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Create(&productWithoutAuditFields{Name: "product1"}))
	})
	t.Run("update", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `products` SET `name` = ?, `updated_by` = ?")).
			WithArgs("product1", "user@test.com").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Model(&product{}).Update(&product{Name: "product1"}))
	})
	t.Run("update with no user", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `products` SET `name` = ?")).
			WithArgs("product1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		assert.NotNil(t, gormDB.Model(&product{}).Update(&product{Name: "product1"}))
	})
	t.Run("update with productWithoutAuditFields", func(t *testing.T) {
		gormDB, mock, defeFunc := getDB(t)
		defer defeFunc()
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE `productWithoutAuditFields` SET `name` = ?")).
			WithArgs("product1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectClose()
		gormDB = gormDB.Set("audit:current_user", "user@test.com")
		assert.NotNil(t, gormDB.Model(&productWithoutAuditFields{}).Update(&productWithoutAuditFields{Name: "product1"}))
	})
}

func getDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("mysql", db)
	assert.Nil(t, err)

	f := func() {
		_ = gormDB.Close()
		_ = db.Close()
	}

	RegisterAuditCallbacks(gormDB)
	return gormDB, mock, f
}
