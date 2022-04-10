package dbtables

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/tucats/ego/app-cli/settings"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/util"
)

// This defines a single operation performed as part of a transaction
type TxOperation struct {
	Opcode  string                 `json:"operation"`
	Table   string                 `json:"table"`
	Filters []string               `json:"filters,omitempty"`
	Columns []string               `json:"columns,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// DeleteRows deletes rows from a table. If no filter is provided, then all rows are
// deleted and the tale is empty. If filter(s) are applied, only the matching rows
// are deleted. The function returns the number of rows deleted.
func Transaction(user string, isAdmin bool, sessionID int32, w http.ResponseWriter, r *http.Request) {
	if e := util.AcceptedMediaType(r, []string{defs.RowCountMediaType}); !errors.Nil(e) {
		util.ErrorResponse(w, sessionID, e.Error(), http.StatusBadRequest)

		return
	}

	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.FilterParameterName: defs.Any,
		defs.UserParameterName:   datatypes.StringTypeName,
	}); !errors.Nil(invalid) {
		util.ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	// Validate the transaction payload.
	tasks := []TxOperation{}
	e := json.NewDecoder(r.Body).Decode(&tasks)
	if e != nil {
		util.ErrorResponse(w, sessionID, "transaction request decode error; "+e.Error(), http.StatusBadRequest)

		return
	}

	ui.Debug(ui.ServerLogger, "[%d] Transaction request with %d operations", sessionID, len(tasks))

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	if len(tasks) == 0 {
		ui.Debug(ui.ServerLogger, "[%d] no tasks in transaction", sessionID)
		w.WriteHeader(200)
		w.Write([]byte("no tasks in transaction"))

		return
	}

	for n, task := range tasks {
		opcode := strings.ToUpper(task.Opcode)
		if !util.InList(opcode, "DELETE", "UPDATE", "INSERT", "DROP") {
			msg := fmt.Sprintf("transaction operation %d has invalid opcode: %s",
				n, opcode)
			util.ErrorResponse(w, sessionID, msg, http.StatusBadRequest)

			return
		}
	}

	// Access the database and execute the transaction operations
	rowsAffected := 0

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		defer db.Close()

		tx, err := db.Begin()
		if !errors.Nil(err) {
			util.ErrorResponse(w, sessionID, "unable to start transaction; "+err.Error(), http.StatusInternalServerError)

			return
		}

		for n, task := range tasks {
			var opErr error

			tableName, _ := fullName(user, task.Table)
			ui.Debug(ui.TableLogger, "[%d] operation %s on table %s", sessionID, task.Opcode, tableName)

			switch strings.ToUpper(task.Opcode) {
			case "UPDATE":
				count := 0
				count, opErr = txUpdate(w, r, sessionID, user, db, tx, task)
				rowsAffected += count

			case "DELETE":
				count := 0
				count, opErr = txDelete(w, r, sessionID, user, tx, task)
				rowsAffected += count

			case "INSERT":
				opErr = txInsert(w, r, sessionID, user, db, tx, task)

			case "DROP":
				opErr = txDrop(w, r, sessionID, user, db, task)
			}

			if !errors.Nil(opErr) {
				_ = tx.Rollback()
				msg := fmt.Sprintf("transaction rollback after %d operations; %s", n+1, opErr.Error())

				util.ErrorResponse(w, sessionID, msg, http.StatusInternalServerError)

				return
			}
		}

		err = tx.Commit()
		if err != nil {
			util.ErrorResponse(w, sessionID, "transaction commit error; "+err.Error(), http.StatusInternalServerError)

			return
		}

		r := defs.DBRowCount{
			ServerInfo: util.MakeServerInfo(sessionID),
			Count:      rowsAffected,
		}
		b, _ := json.MarshalIndent(r, ui.JSONIndentPrefix, ui.JSONIndentSpacer)

		ui.Debug(ui.TableLogger, "[%d] %s",
			sessionID,
			fmt.Sprintf("completed %d operations in transaction, updated %d rows", len(tasks), rowsAffected))

		w.Header().Add("Content-Type", defs.RowCountMediaType)
		w.WriteHeader(http.StatusOK)
		w.Write(b)

		return
	}
}

func txUpdate(w http.ResponseWriter, r *http.Request, sessionID int32, user string, db *sql.DB, tx *sql.Tx, task TxOperation) (int, error) {
	tableName, _ := fullName(user, task.Table)

	validColumns, err := getColumnInfo(db, user, tableName, sessionID)
	if !errors.Nil(err) {
		msg := "Unable to read table metadata, " + err.Error()

		return 0, errors.NewMessage(msg)
	}

	// Make sure none of the columns in the update are non-existant
	for k := range task.Data {
		valid := false
		for _, column := range validColumns {
			if column.Name == k {
				valid = true

				break
			}
		}

		if !valid {
			msg := "insert task refernces non-existant column: " + k

			return 0, errors.NewMessage(msg)
		}
	}

	// Is there columns list for this task that should be used to determine
	// which parts of the payload to use?
	if len(task.Columns) > 0 {
		// Make sure none of the columns in the columns are non-existant
		for _, name := range task.Columns {
			valid := false
			for _, k := range validColumns {
				if name == k.Name {
					valid = true

					break
				}
			}

			if !valid {
				msg := "insert task references non-existant column: " + name

				return 0, errors.NewMessage(msg)
			}
		}

		// The columns list is valid, so use it to thin out the task payload
		keepList := map[string]bool{}

		for k := range task.Data {
			keepList[k] = false
		}

		for _, columnName := range task.Columns {
			keepList[columnName] = true
		}

		for k, keep := range keepList {
			if !keep {
				delete(task.Data, k)
			}
		}
	}

	// Form the update query. We start with a list of the keys to update
	// in a predictable order
	var result strings.Builder

	var values []interface{}

	var keys []string

	for key := range task.Data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	result.WriteString("UPDATE ")
	result.WriteString(tableName)

	// Loop over the item names and add SET clauses for each one. We always
	// ignore the rowid value because you cannot update it on an UPDATE call;
	// it is only set on an insert.
	columnPosition := 0

	for _, key := range keys {
		if key == defs.RowIDName {
			continue
		}

		// Add the value to the list of values that will be passed to the Exec()
		// function later. These must be in the same order that the column names
		// are specified in the query text.
		values = append(values, task.Data[key])

		if columnPosition == 0 {
			result.WriteString(" SET ")
		} else {
			result.WriteString(", ")
		}

		columnPosition++

		result.WriteString("\"" + key + "\"")
		result.WriteString(fmt.Sprintf(" = $%d", columnPosition))
	}

	// If there is a filter, then add that as well. And fail if there
	// isn't a filter but must be
	if filter := whereClause(task.Filters); filter != "" {
		result.WriteString(filter)
	} else if settings.GetBool(defs.TablesServerEmptyFilterError) {
		ui.Debug(ui.ServerLogger, "DEBUG: filters = %v", task.Filters)
		return 0, errors.NewMessage("update without filter is not allowed")
	}

	ui.Debug(ui.TableLogger, "[%d] Query: ", sessionID, result.String())

	queryResult, updateErr := tx.Exec(result.String(), values...)
	count, _ := queryResult.RowsAffected()

	return int(count), errors.New(updateErr)
}

func txDelete(w http.ResponseWriter, r *http.Request, sessionID int32, user string, tx *sql.Tx, task TxOperation) (int, error) {

	tableName, _ := fullName(user, task.Table)

	ui.Debug(ui.ServerLogger, "[%d] transaction task to delete rows from table %s", sessionID, tableName)

	if p := parameterString(r); p != "" {
		ui.Debug(ui.ServerLogger, "[%d] request parameters:  %s", sessionID, p)
	}

	if where := whereClause(task.Filters); where == "" {
		if settings.GetBool(defs.TablesServerEmptyFilterError) {
			return 0, errors.NewMessage("operation invalid with empty filter")
		}
	}

	q := formSelectorDeleteQuery(r.URL, task.Filters, "", tableName, user, deleteVerb)
	if p := strings.Index(q, syntaxErrorPrefix); p >= 0 {
		return 0, errors.NewMessage(filterErrorMessage(q))
	}

	ui.Debug(ui.TableLogger, "[%d] Exec: %s", sessionID, q)

	rows, err := tx.Exec(q)
	if err == nil {
		rowCount, _ := rows.RowsAffected()
		if rowCount == 0 && settings.GetBool(defs.TablesServerEmptyRowsetError) {
			return 0, errors.NewMessage("no matching rows")
		}

		ui.Debug(ui.TableLogger, "[%d] Deleted %d rows; %d", sessionID, rowCount, 200)

		count, _ := rows.RowsAffected()

		return int(count), nil
	}

	return 0, err
}

func txDrop(w http.ResponseWriter, r *http.Request, sessionID int32, user string, db *sql.DB, task TxOperation) error {

	table, _ := fullName(user, task.Table)

	q := "DROP TABLE " + table
	_, err := db.Exec(q)

	return err
}

func txInsert(w http.ResponseWriter, r *http.Request, sessionID int32, user string, db *sql.DB, tx *sql.Tx, task TxOperation) error {
	// Get the column metadata for the table we're insert into, so we can validate column info.
	tableName, _ := fullName(user, task.Table)

	columns, err := getColumnInfo(db, user, tableName, sessionID)
	if !errors.Nil(err) {
		return errors.NewMessage("unable to read table metadata; " + err.Error())
	}

	// If we're showing our payload in the log, do that now
	if ui.LoggerIsActive(ui.RestLogger) {
		b, _ := json.MarshalIndent(task.Data, ui.JSONIndentPrefix, ui.JSONIndentSpacer)

		ui.Debug(ui.RestLogger, "[%d] INSERT task payload:\n%s", sessionID, util.SessionLog(sessionID, string(b)))
	}

	// It's a new row, so assign a UUID now. This overrides any previous item in the payload
	// for _row_id_ or creates it if not found. Row IDs are always assigned on insert only.
	task.Data[defs.RowIDName] = uuid.New().String()

	for _, column := range columns {
		v, ok := task.Data[column.Name]
		if !ok && settings.GetBool(defs.TableServerPartialInsertError) {
			expectedList := make([]string, 0)
			for _, k := range columns {
				expectedList = append(expectedList, k.Name)
			}

			providedList := make([]string, 0)
			for k := range task.Data {
				providedList = append(providedList, k)
			}

			sort.Strings(expectedList)
			sort.Strings(providedList)

			msg := fmt.Sprintf("Payload did not include data for \"%s\"; expected %v but payload contained %v",
				column.Name, strings.Join(expectedList, ","), strings.Join(providedList, ","))
			return errors.NewMessage(msg)
		}

		// If it's one of the date/time values, make sure it is wrapped in single qutoes.
		if keywordMatch(column.Type, "time", "date", "timestamp") {
			text := strings.TrimPrefix(strings.TrimSuffix(datatypes.GetString(v), "\""), "\"")
			task.Data[column.Name] = "'" + strings.TrimPrefix(strings.TrimSuffix(text, "'"), "'") + "'"
			ui.Debug(ui.TableLogger, "[%d] updated column %s value from %v to %v", sessionID, column.Name, v, task.Data[column.Name])
		}
	}

	q, values := formInsertQuery(task.Table, user, task.Data)
	ui.Debug(ui.TableLogger, "[%d] Insert row with query: %s", sessionID, q)

	_, e := tx.Exec(q, values...)
	if e != nil {
		return errors.NewMessage("error inserting row; " + e.Error())
	}

	ui.Debug(ui.TableLogger, "[%d] successful INSERT to %s", sessionID, tableName)

	return nil
}
