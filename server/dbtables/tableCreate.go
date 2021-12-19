package dbtables

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/defs"
)

// TableCreate creates a new table based on the JSON payload
func TableCreate(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	var err error

	db, err := OpenDB(sessionID, user, "")

	if err == nil && db != nil {

		if !isAdmin && Authorized(sessionID, nil, user, tableName, updateOperation) {
			ErrorResponse(w, sessionID, "User does not have update permission", http.StatusForbidden)
			return
		}

		data := []defs.DBColumn{}

		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			ErrorResponse(w, sessionID, "Invalid table create payload: "+err.Error(), http.StatusBadRequest)

			return
		}

		q := formCreateQuery(r.URL, user, isAdmin, data, sessionID, w)
		if q == "" {
			return
		}

		if !createSchemaIfNeeded(w, sessionID, db, user, tableName) {
			return
		}

		ui.Debug(ui.ServerLogger, "[%d] Query: %s", sessionID, q)

		counts, err := db.Exec(q)
		if err == nil {
			rows, _ := counts.RowsAffected()
			result := defs.DBRowCount{Count: int(rows), RestResponse: defs.RestResponse{Status: 200}}

			CreateTablePermissions(sessionID, db, user, tableName, readOperation, deleteOperation, updateOperation)

			b, _ := json.MarshalIndent(result, "", "  ")
			_, _ = w.Write(b)
			ui.Debug(ui.ServerLogger, "[%d] table created", sessionID)

			return
		}

		ui.Debug(ui.ServerLogger, "[%d] Error creating table, %v", sessionID, err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))

		return
	}

	ui.Debug(ui.ServerLogger, "[%d] Error inserting into table, %v", sessionID, err)
	w.WriteHeader(http.StatusBadRequest)
	if err == nil {
		err = fmt.Errorf("unknown error")
	}
	_, _ = w.Write([]byte(err.Error()))

}

// Verify that the schema exists for this user, and create it if not found.
func createSchemaIfNeeded(w http.ResponseWriter, sessionID int32, db *sql.DB, user string, tableName string) bool {

	schema := user
	if dot := strings.Index(tableName, "."); dot >= 0 {
		schema = tableName[:dot]
	}

	q := queryParameters(createSchemaString, map[string]string{
		"schema": schema,
	})

	result, err := db.Exec(q)
	if err != nil {
		ErrorResponse(w, sessionID, "Error creating schema; "+err.Error(), http.StatusInternalServerError)
		return false
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		ui.Debug(ui.ServerLogger, "[%d] Created schema %s", sessionID, schema)
	}

	return true
}
