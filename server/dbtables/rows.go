package dbtables

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/util"
)

// DeleteRows deletes rows from a table. If no filter is provided, then all rows are
// deleted and the tale is empty. If filter(s) are applied, only the matching rows
// are deleted. The function returns the number of rows deleted.
func DeleteRows(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	tableName, _ = fullName(user, tableName)
	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.FilterParameterName: defs.Any,
		defs.UserParameterName:   "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	ui.Debug(ui.ServerLogger, "[%d] Request to delete rows from table %s", sessionID, tableName)

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		if !isAdmin && Authorized(sessionID, nil, user, tableName, deleteOperation) {
			ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return
		}

		q := formSelectorDeleteQuery(r.URL, user, deleteVerb)
		if p := strings.Index(q, syntaxErrorPrefix); p > 0 {
			ErrorResponse(w, sessionID, filterErrorMessage(q), http.StatusBadRequest)

			return
		}

		ui.Debug(ui.TableLogger, "[%d] Exec: %s", sessionID, q)

		rows, err := db.Exec(q)
		if err == nil {
			rowCount, _ := rows.RowsAffected()

			resp := defs.DBRowCount{
				Count: int(rowCount),
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			_, _ = w.Write(b)

			ui.Debug(ui.TableLogger, "[%d] Deleted %d rows ", sessionID, rowCount)

			return
		}
	}

	ui.Debug(ui.ServerLogger, "[%d] Error deleting from table, %v", sessionID, err)
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(err.Error()))
}

// InsertRows updates the rows (specified by a filter clause as needed) with the data from the payload.
func InsertRows(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	var err error

	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	tableName, _ = fullName(user, tableName)

	ui.Debug(ui.ServerLogger, "[%d] Request to insert rows into table %s", sessionID, tableName)

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		// Note that "update" here means add to or change the row. So we check "update"
		// on test for insert permissions
		if !isAdmin && Authorized(sessionID, nil, user, tableName, updateOperation) {
			ErrorResponse(w, sessionID, "User does not have insert permission", http.StatusForbidden)

			return
		}

		// Get the column metadata for the table we're insert into, so we can validate column info.
		var columns []defs.DBColumn

		tableName, _ = fullName(user, tableName)

		columns, err = getColumnInfo(db, user, tableName, sessionID)
		if !errors.Nil(err) {
			ErrorResponse(w, sessionID, "Unable to read table metadata, "+err.Error(), http.StatusBadRequest)

			return
		}

		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		rawPayload := buf.String()

		ui.Debug(ui.RestLogger, "[%d] RAW payload:\n%s", sessionID, rawPayload)

		// Lets get the rows we are to insert. This is either a row set, or a single object.
		rowSet := defs.DBRows{}

		err = json.Unmarshal([]byte(rawPayload), &rowSet)
		if err != nil || len(rowSet.Rows) == 0 {
			// Not a valid row set, but might be a single item
			item := map[string]interface{}{}

			err = json.Unmarshal([]byte(rawPayload), &item)
			if err != nil {
				ErrorResponse(w, sessionID, "Invalid INSERT payload: "+err.Error(), http.StatusBadRequest)

				return
			} else {
				rowSet.Count = 1
				rowSet.Rows = make([]map[string]interface{}, 1)
				rowSet.Rows[0] = item
				ui.Debug(ui.RestLogger, "[%d] Converted object to rowset payload %v", sessionID, item)
			}
		} else {
			ui.Debug(ui.RestLogger, "[%d] Received rowset with %d items", sessionID, len(rowSet.Rows))
		}

		// If we're showing our payload in the log, do that now
		if ui.LoggerIsActive(ui.RestLogger) {
			b, _ := json.MarshalIndent(rowSet, ui.JSONIndentPrefix, ui.JSONIndentSpacer)

			ui.Debug(ui.RestLogger, "[%d] Resolved REST Request payload:\n%s", sessionID, string(b))
		}

		// If at this point we have an empty row set, then just bail out now. Return a success
		// status but an indicator that nothing was done.
		if len(rowSet.Rows) == 0 {
			ErrorResponse(w, sessionID, "No rows found in INSERT payload", http.StatusNoContent)

			return
		}

		// For any object in the payload, we must assign a UUID now. This overrides any previous
		// item in the set for _row_id_ or creates it if not found. Row IDs are always assigned
		// on input only.
		for n := 0; n < len(rowSet.Rows); n++ {
			rowSet.Rows[n][defs.RowIDName] = uuid.New().String()
		}

		// Start a transaction, and then lets loop over the rows in the rowset. Note this might
		// be just one row.
		tx, _ := db.Begin()
		count := 0

		for _, row := range rowSet.Rows {
			for _, column := range columns {
				v, ok := row[column.Name]
				if !ok {
					ErrorResponse(w, sessionID, "Invalid column in request payload: "+column.Name, http.StatusBadRequest)

					return
				}

				// If it's one of the date/time values, make sure it is wrapped in single qutoes.
				if keywordMatch(column.Type, "time", "date", "timestamp") {
					text := strings.TrimPrefix(strings.TrimSuffix(datatypes.GetString(v), "\""), "\"")
					row[column.Name] = "'" + strings.TrimPrefix(strings.TrimSuffix(text, "'"), "'") + "'"
					ui.Debug(ui.TableLogger, "[%d] updated column %s value from %v to %v", sessionID, column.Name, v, row[column.Name])
				}
			}

			q, values := formInsertQuery(r.URL, user, row)
			ui.Debug(ui.TableLogger, "[%d] Insert row with query: %s", sessionID, q)

			_, err := db.Exec(q, values...)
			if err == nil {
				count++
			} else {
				ErrorResponse(w, sessionID, err.Error(), http.StatusConflict)
				_ = tx.Rollback()

				return
			}
		}

		if err == nil {
			result := defs.DBRowCount{
				Count: count,
			}

			b, _ := json.MarshalIndent(result, "", "  ")
			_, _ = w.Write(b)

			err = tx.Commit()
			if err == nil {
				ui.Debug(ui.TableLogger, "[%d] Inserted %d rows", sessionID, count)

				return
			}
		}

		_ = tx.Rollback()

		ErrorResponse(w, sessionID, "insert error: "+err.Error(), http.StatusInternalServerError)

		return
	}

	if !errors.Nil(err) {
		ErrorResponse(w, sessionID, "insert error: "+err.Error(), http.StatusInternalServerError)
	}
}

// ReadRows reads the data for a given table, and returns it as an array
// of structs for each row, with the struct tag being the column name. The
// query can also specify filter, sort, and column query parameters to refine
// the read operation.
func ReadRows(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	tableName, _ = fullName(user, tableName)

	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.StartParameterName:  "int",
		defs.LimitParameterName:  "int",
		defs.ColumnParameterName: "list",
		defs.SortParameterName:   "list",
		defs.FilterParameterName: defs.Any,
		defs.UserParameterName:   "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	ui.Debug(ui.ServerLogger, "[%d] Request to read rows from table %s", sessionID, tableName)

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		if !isAdmin && Authorized(sessionID, nil, user, tableName, readOperation) {
			ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return
		}

		q := formSelectorDeleteQuery(r.URL, user, selectVerb)
		if p := strings.Index(q, syntaxErrorPrefix); p > 0 {
			ErrorResponse(w, sessionID, filterErrorMessage(q), http.StatusBadRequest)

			return
		}

		ui.Debug(ui.TableLogger, "[%d] Query: %s", sessionID, q)

		err = readRowData(db, q, sessionID, w)
		if err == nil {
			return
		}
	}

	ui.Debug(ui.TableLogger, "[%d] Error reading table, %v", sessionID, err)
	ErrorResponse(w, sessionID, err.Error(), 400)
}

func readRowData(db *sql.DB, q string, sessionID int32, w http.ResponseWriter) error {
	var rows *sql.Rows

	var err error

	result := []map[string]interface{}{}
	rowCount := 0

	rows, err = db.Query(q)
	if err == nil {
		defer rows.Close()

		columnNames, _ := rows.Columns()
		columnCount := len(columnNames)

		for rows.Next() {
			row := make([]interface{}, columnCount)
			rowptrs := make([]interface{}, columnCount)

			for i := range row {
				rowptrs[i] = &row[i]
			}

			err = rows.Scan(rowptrs...)
			if err == nil {
				newRow := map[string]interface{}{}
				for i, v := range row {
					newRow[columnNames[i]] = v
				}

				result = append(result, newRow)
				rowCount++
			}
		}

		resp := defs.DBRows{
			Rows:  result,
			Count: len(result),
		}

		b, _ := json.MarshalIndent(resp, "", "  ")
		_, _ = w.Write(b)

		ui.Debug(ui.TableLogger, "[%d] Read %d rows of %d columns", sessionID, rowCount, columnCount)
	}

	return err
}

// UpdateRows updates the rows (specified by a filter clause as needed) with the data from the payload.
func UpdateRows(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	tableName, _ = fullName(user, tableName)
	count := 0

	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.FilterParameterName: defs.Any,
		defs.UserParameterName:   "string",
		defs.ColumnParameterName: "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	ui.Debug(ui.ServerLogger, "[%d] Request to update rows in table %s", sessionID, tableName)

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		if !isAdmin && Authorized(sessionID, nil, user, tableName, updateOperation) {
			ErrorResponse(w, sessionID, "User does not have update permission", http.StatusForbidden)

			return
		}

		excludeList := map[string]bool{}

		p := r.URL.Query()
		if v, found := p[defs.ColumnParameterName]; found {
			// There is a column list, so build a list of all the columns, and then
			// remove the ones from the column parameter. This builds a list of columns
			// that are excluded.
			columns, err := getColumnInfo(db, user, tableName, sessionID)
			if !errors.Nil(err) {
				ErrorResponse(w, sessionID, err.Error(), http.StatusInternalServerError)

				return
			}

			for _, column := range columns {
				excludeList[column.Name] = true
			}

			for _, name := range v {
				nameParts := strings.Split(stripQuotes(name), ",")
				for _, part := range nameParts {
					if part != "" {
						excludeList[part] = false
					}
				}
			}
		}

		// For debugging, show the raw payload. We may remove this later...
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, r.Body)
		rawPayload := buf.String()

		ui.Debug(ui.RestLogger, "[%d] RAW payload:\n%s", sessionID, rawPayload)

		// Lets get the rows we are to update. This is either a row set, or a single object.
		rowSet := defs.DBRows{}

		err = json.Unmarshal([]byte(rawPayload), &rowSet)
		if err != nil || len(rowSet.Rows) == 0 {
			// Not a valid row set, but might be a single item
			item := map[string]interface{}{}

			err = json.Unmarshal([]byte(rawPayload), &item)
			if err != nil {
				ErrorResponse(w, sessionID, "Invalid UPDATE payload: "+err.Error(), http.StatusBadRequest)

				return
			} else {
				rowSet.Count = 1
				rowSet.Rows = make([]map[string]interface{}, 1)
				rowSet.Rows[0] = item
				ui.Debug(ui.RestLogger, "[%d] Converted object to rowset payload %v", sessionID, item)
			}
		} else {
			ui.Debug(ui.RestLogger, "[%d] Received rowset with %d items", sessionID, len(rowSet.Rows))
		}

		// Anything in the data map that is on the exclude list is removed
		ui.Debug(ui.TableLogger, "[%d] exclude list = %v", sessionID, excludeList)

		// Start a transaction to ensure atomicity of the entire update
		tx, _ := db.Begin()

		// Loop over the row set doing the insert
		for _, data := range rowSet.Rows {
			hasRowID := false

			if v, found := data[defs.RowIDName]; found {
				if datatypes.GetString(v) != "" {
					hasRowID = true
				}
			}

			for key, excluded := range excludeList {
				if key == defs.RowIDName && hasRowID {
					continue
				}

				if excluded {
					delete(data, key)
				}
			}

			ui.Debug(ui.TableLogger, "[%d] values list = %v", sessionID, data)

			q, values := formUpdateQuery(r.URL, user, data)
			if p := strings.Index(q, syntaxErrorPrefix); p > 0 {
				ErrorResponse(w, sessionID, filterErrorMessage(q), http.StatusBadRequest)

				return
			}

			ui.Debug(ui.TableLogger, "[%d] Query: %s", sessionID, q)

			counts, err := db.Exec(q, values...)
			if err == nil {
				rowsAffected, _ := counts.RowsAffected()
				count = count + int(rowsAffected)
			} else {
				ErrorResponse(w, sessionID, err.Error(), http.StatusConflict)
				_ = tx.Rollback()

				return
			}
		}

		if errors.Nil(err) {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}

	if errors.Nil(err) {
		result := defs.DBRowCount{
			Count: count,
		}

		b, _ := json.MarshalIndent(result, "", "  ")
		_, _ = w.Write(b)

		ui.Debug(ui.TableLogger, "[%d] Updated %d rows", sessionID, count)
	} else {
		ErrorResponse(w, sessionID, "Error updating table, "+err.Error(), http.StatusInternalServerError)
	}
}

func filterErrorMessage(q string) string {
	if p := strings.Index(q, syntaxErrorPrefix); p > 0 {
		msg := q[p+len(syntaxErrorPrefix):]
		if p := strings.Index(msg, defs.RowIDName); p > 0 {
			msg = msg[:p]
		}

		return "filter error: " + msg
	}

	return q
}
