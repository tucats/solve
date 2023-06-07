package resources

import (
	"reflect"

	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/data"
)

func (r *ResHandle) Read(filters ...*Filter) ([]interface{}, error) {
	var (
		err error
	)

	if r.Database == nil {
		return nil, ErrDatabaseNotOpen
	}

	var results []interface{}

	sql := r.readRowSQL()

	for index, filter := range filters {
		if index == 0 {
			sql = sql + " where "
		} else {
			sql = sql + " and "
		}

		sql = sql + filter.Generate()
	}

	// Add any active order-by clause
	sql = sql + r.OrderBy()

	ui.Log(ui.DBLogger, "[0] Resource read: %s", sql)

	rows, err := r.Database.Query(sql)
	if rows != nil {
		defer rows.Close()
	}

	count := 0

	if err == nil {
		for rows.Next() {
			rowData := make([]interface{}, len(r.Columns))
			rowDataPtrs := make([]interface{}, len(r.Columns))

			for i := range rowDataPtrs {
				rowDataPtrs[i] = &rowData[i]
			}

			err = rows.Scan(rowDataPtrs...)

			if err == nil {
				value := reflect.New(r.Type).Interface()
				count++

				for i := 0; i < len(rowData); i++ {
					switch r.Columns[i].SQLType {
					case "integer":
						reflect.ValueOf(value).Elem().Field(i).SetInt(data.Int64(rowData[i]))
					case "boolean":
						reflect.ValueOf(value).Elem().Field(i).SetBool(data.Bool(rowData[i]))
					case "char varying":
						reflect.ValueOf(value).Elem().Field(i).SetString(data.String(rowData[i]))
					}
				}

				results = append(results, value)
			}
		}
	}

	if err == nil {
		ui.Log(ui.DBLogger, "[0] Resource list read %d rows", count)
	}

	return results, err
}