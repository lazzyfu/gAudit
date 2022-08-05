/*
@Time    :   2022/06/23 16:37:41
@Author  :   zongfei.fu
@Desc    :   操作目标审核数据库
*/

package utils

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

// DB Struct
type DB struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
	Timeout  time.Duration
}

// Open connection for db
func (d *DB) Open() (*sql.DB, error) {
	// dbms_monitor:1234.com@tcp(127.0.0.1:3306)/noah_db
	config := mysql.Config{
		User:                 d.User,
		Passwd:               d.Password,
		Addr:                 fmt.Sprintf("%s:%d", d.Host, d.Port),
		Net:                  "tcp",
		DBName:               d.Database,
		AllowNativePasswords: true,
		Timeout:              d.Timeout * time.Millisecond,
		ReadTimeout:          d.Timeout * time.Millisecond,
		WriteTimeout:         d.Timeout * time.Millisecond,
	}

	DSN := config.FormatDSN()
	db, err := sql.Open("mysql", DSN)
	return db, err
}

// Executes a query without returning any rows.
func (d *DB) Exec(query string) error {
	db, err := d.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(query)
	return err
}

// FetchRows
func (d *DB) FetchRows(query string) (*[]map[string]interface{}, error) {
	db, err := d.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	// Get column names
	columns, error := rows.Columns()
	if error != nil {
		return nil, error
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	resultSlice := make([]map[string]interface{}, 0)
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		var value string
		vmap := make(map[string]interface{}, len(scanArgs))

		for i, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			vmap[columns[i]] = value
		}
		resultSlice = append(resultSlice, vmap)
	}
	return &resultSlice, nil
}
