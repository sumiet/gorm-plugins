package upsertutility

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"

	"github.com/jinzhu/gorm"
)

const (
	// Query Constants
	_insertInto           = "INSERT INTO "
	_insertIgnoreInto     = "INSERT IGNORE INTO "
	_values               = "VALUES"
	_onDuplicateKeyUpdate = "ON DUPLICATE KEY UPDATE"
	_openingBracket       = "("
	_closingBracket       = ")"
	_comma                = ", "
	_updatedByObjectFieldName = "updatedBy"
	_createdByObjectFieldName = "createdBy"

	// maxPlaceholderSize is the maximum number of placeholders one bulk query can contain
	_maxPlaceholderSize = 10000
)

// BulkUpsertStats contains the stats for bulk execution
type BulkUpsertStats struct {
	UpsertCount     int
	UpsertFailCount int
}

// BatchUpsert inserts into mysql by creating a bulk insert query
func BatchUpsert(db *gorm.DB, objArr []interface{}, ignoreErrsFlag bool) (*BulkUpsertStats, error) {
	// If there is no data, nothing to do.
	if len(objArr) == 0 {
		return nil, nil
	}

	// object to track success and failure counts
	bulkUpsertStats := BulkUpsertStats{}

	// prepare main object
	mainObj := objArr[0]
	mainScope := db.NewScope(mainObj)
	colList, colUpdateList := getColumnLists(mainScope)

	// using the audit plugin from: https://github.com/sumiet/gorm-plugins/tree/master/audit
	user, _ := audit.GetCurrentUser(mainScope)

	// total input size
	lengthOfInput := len(objArr)
	// number of columns in the table on which bulk insert is being done
	numberOfFields := len(colList)

	// since mysql has limitation on number of placeholders, there is only
	// a fixed size of batch which can be inserted in one shot
	// roughly 1000 for object with 60 columns
	eachBatchSize := _maxPlaceholderSize / numberOfFields
	// based on the batch size we figure out the number of batches in which
	// data has to be broken to insert
	numberOfBatches := int(math.Ceil(float64(lengthOfInput) / float64(eachBatchSize)))
	// looping to process, based on number of batches to be executed
	for index := 0; index < numberOfBatches; index++ {
		// resetting main scope SQL vars to empty slice
		mainScope.SQLVars = mainScope.SQLVars[:0]
		startPosition := index * eachBatchSize
		endPosition := startPosition + eachBatchSize
		// edge case when the final set has lesser records than expected
		if endPosition > lengthOfInput {
			endPosition = lengthOfInput
		}
		thisBatchSize := endPosition - startPosition
		placeholdersArr := getPlaceholdersArray(objArr[startPosition:endPosition], db, mainScope, user)
		// executing upsert
		if _, err := executeUpsert(mainScope, colList, colUpdateList, placeholdersArr, ignoreErrsFlag); err != nil {
			bulkUpsertStats.UpsertFailCount += thisBatchSize
			return &bulkUpsertStats, err
		}
		bulkUpsertStats.UpsertCount = bulkUpsertStats.UpsertCount + thisBatchSize
	}
	return &bulkUpsertStats, nil
}

// getPlaceholdersArray gives the placeholders array for the given data set
func getPlaceholdersArray(objArr []interface{}, db *gorm.DB, mainScope *gorm.Scope, user string) []string {
	var placeholdersArr []string

	for _, obj := range objArr {
		scope := db.NewScope(obj)

		fields := scope.Fields()

		placeholders := make([]string, 0, len(fields))
		for i := range fields {
			// If primary key has blank value (0 for int, "" for string, nil for interface ...), skip it.
			// If field is ignore field, skip it.
			// If a default value field has blank value (0 for int, "" for string, nil for interface ...), skip it.
			if (fields[i].IsPrimaryKey && fields[i].IsBlank) ||
				(fields[i].IsIgnored) ||
				(fields[i].HasDefaultValue && fields[i].IsBlank) {
				continue
			}

			placeholders = append(placeholders, scope.AddToVars(getFieldValue(user, fields[i])))
		}
		placeholdersStr := fmt.Sprintf("%s%s%s",
			_openingBracket,
			strings.Join(placeholders, _comma),
			_closingBracket)
		placeholdersArr = append(placeholdersArr, placeholdersStr)
		// add real variables for the replacement of placeholders' '?' letter later.
		mainScope.SQLVars = append(mainScope.SQLVars, scope.SQLVars...)
	}
	return placeholdersArr
}

// getFieldValue gives the field value as per the column name
func getFieldValue(user string, field *gorm.Field) interface{} {
	switch field.Name {
	case _updatedByObjectFieldName, _createdByObjectFieldName:
		return user
	default:
		return field.Field.Interface()
	}
}

// getColumnLists returns the list of columns to used for query string formation
func getColumnLists(mainScope *gorm.Scope) ([]string, []string) {
	var colList, colUpdateList []string

	// getting the fields from the objects
	mainFields := mainScope.Fields()
	for i := range mainFields {
		// If primary key has blank value (0 for int, "" for string, nil for interface ...), skip it.
		// If field is ignore field, skip it.
		// If a default value field has blank value (0 for int, "" for string, nil for interface ...), skip it.
		if (mainFields[i].IsPrimaryKey && mainFields[i].IsBlank) ||
			(mainFields[i].IsIgnored) ||
			(mainFields[i].HasDefaultValue && mainFields[i].IsBlank) {
			continue
		}
		colName := mainScope.Quote(mainFields[i].DBName)
		colList = append(colList, colName)
		colUpdateList = append(colUpdateList, fmt.Sprintf("%s=%s%s%s%s", colName, _values, _openingBracket, colName, _closingBracket))
	}

	return colList, colUpdateList
}

// executeUpsert prepares the raw SQL and executes the insert statement in the DB
func executeUpsert(mainScope *gorm.Scope, colList, colUpdateList, placeholdersArr []string, ignoreErrsFlag bool) (sql.Result, error) {

	// Find the 'Insert' command based on 'ignoreErrs' flag
	insertInto := _insertInto
	if ignoreErrsFlag {
		insertInto = _insertIgnoreInto
	}

	mainScope.Raw(fmt.Sprintf("%s %s %s%s%s %s %s %s %s",
		insertInto,
		mainScope.QuotedTableName(),
		_openingBracket,
		strings.Join(colList, _comma),
		_closingBracket,
		_values,
		strings.Join(placeholdersArr, _comma),
		_onDuplicateKeyUpdate,
		strings.Join(colUpdateList, _comma),
	))

	// executing insert
	return mainScope.SQLDB().Exec(mainScope.SQL, mainScope.SQLVars...)
}
