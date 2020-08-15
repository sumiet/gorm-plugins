# Audit

Audit is used to record the last User who created and/or updated your [GORM](https://github.com/jinzhu/gorm) model. It does so using a `CreatedBy` and `UpdatedBy` field and defaulting the `CreatedAt` and `UpdatedAt` time fields to `CURRENT_TIMESTAMP`

### Register GORM Callbacks

Audit utilizes [GORM](https://github.com/jinzhu/gorm) callbacks to log data, so you will need to register callbacks first:

```go
  // register db callbacks by using your *gorm.DB(dbConn) object like this
  audit.RegisterAuditCallbacks(dbConn)
```

Sample GORM Object:
```go
type MyTable struct {
    
    // table fields
    audit.Model
}
```

Read [here](https://medium.com/@sumit_agarwal/gorm-managing-audit-fields-b8ad8497ba16) from more.

## License

Released under the [MIT License](http://opensource.org/licenses/MIT).
