package tables

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/data"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/util"
)

const unexpectedNilPointerError = "Unexpected nil database object pointer"

// TableCreate creates a new table based on the JSON payload, which must be an array of DBColumn objects, defining
// the characteristics of each column in the table. If the table name is the special name "@sql" the payload instead
// is assumed to be a JSON-encoded string containing arbitrary SQL to exectue. Only an admin user can use the "@sql"
// table name.
func TableCreate(user string, isAdmin bool, tableName string, sessionID int, w http.ResponseWriter, r *http.Request) int {
	var err error

	if err := util.AcceptedMediaType(r, []string{defs.SQLStatementsMediaType, defs.RowSetMediaType, defs.RowCountMediaType}); err != nil {
		return util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)
	}

	// Verify that there are no parameters
	if err := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); err != nil {
		return util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		tableName, _ = fullName(user, tableName)

		if !isAdmin && Authorized(sessionID, db, user, tableName, updateOperation) {
			return util.ErrorResponse(w, sessionID, "User does not have update permission", http.StatusForbidden)
		}

		data := []defs.DBColumn{}

		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			return util.ErrorResponse(w, sessionID, "Invalid table create payload: "+err.Error(), http.StatusBadRequest)
		}

		for _, column := range data {
			if column.Name == "" {
				return util.ErrorResponse(w, sessionID, "Missing or empty column name", http.StatusBadRequest)
			}

			if column.Type == "" {
				return util.ErrorResponse(w, sessionID, "Missing or empty type name", http.StatusBadRequest)
			}

			if !keywordMatch(column.Type, defs.TableColumnTypeNames...) {
				return util.ErrorResponse(w, sessionID, "Invalid type name: "+column.Type, http.StatusBadRequest)
			}
		}

		q := formCreateQuery(r.URL, user, isAdmin, data, sessionID, w)
		if q == "" {
			return http.StatusOK
		}

		if !createSchemaIfNeeded(w, sessionID, db, user, tableName) {
			return http.StatusOK
		}

		ui.Log(ui.SQLLogger, "[%d] Exec: %s", sessionID, q)

		counts, err := db.Exec(q)
		if err == nil {
			rows, _ := counts.RowsAffected()
			result := defs.DBRowCount{
				ServerInfo: util.MakeServerInfo(sessionID),
				Count:      int(rows),
			}

			tableName, _ = fullName(user, tableName)

			CreateTablePermissions(sessionID, db, user, tableName, readOperation, deleteOperation, updateOperation)
			w.Header().Add("Content-Type", defs.RowCountMediaType)

			b, _ := json.MarshalIndent(result, "", "  ")
			_, _ = w.Write(b)

			if ui.IsActive(ui.RestLogger) {
				ui.WriteLog(ui.RestLogger, "[%d] Response payload:\n%s", sessionID, util.SessionLog(sessionID, string(b)))
			}

			ui.Log(ui.ServerLogger, "[%d] table created", sessionID)

			return http.StatusOK
		}

		ui.Log(ui.ServerLogger, "[%d] Error creating table, %v", sessionID, err)

		return util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)
	}

	ui.Log(ui.TableLogger, "[%d] Error inserting into table, %v", sessionID, strings.TrimPrefix(err.Error(), "pq: "))

	if err == nil {
		err = fmt.Errorf("unknown error")
	}

	return util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)
}

// Verify that the schema exists for this user, and create it if not found.
func createSchemaIfNeeded(w http.ResponseWriter, sessionID int, db *sql.DB, user string, tableName string) bool {
	schema := user
	if dot := strings.Index(tableName, "."); dot >= 0 {
		schema = tableName[:dot]
	}

	q := queryParameters(createSchemaQuery, map[string]string{
		"schema": schema,
	})

	result, err := db.Exec(q)
	if err != nil {
		util.ErrorResponse(w, sessionID, "Error creating schema; "+err.Error(), http.StatusInternalServerError)

		return false
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		ui.Log(ui.TableLogger, "[%d] Created schema %s", sessionID, schema)
	}

	return true
}

// ReadTable reads the metadata for a given table, and returns it as an array
// of column names and types.
func ReadTable(user string, isAdmin bool, tableName string, sessionID int, w http.ResponseWriter, r *http.Request) int {
	if err := util.AcceptedMediaType(r, []string{defs.TableMetadataMediaType}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	// Verify that there are no parameters
	if err := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		// Special case; if the table name is @permissions then the payload is processed as request
		// to read all the permissions data
		if strings.EqualFold(tableName, permissionsPseudoTable) {
			if !isAdmin {
				util.ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

				return http.StatusForbidden
			}

			return ReadAllPermissions(db, sessionID, w, r)
		}

		tableName, _ = fullName(user, tableName)

		if !isAdmin && Authorized(sessionID, db, user, tableName, readOperation) {
			util.ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return http.StatusForbidden
		}

		// Determine which columns must be unique
		q := queryParameters(uniqueColumnsQuery, map[string]string{
			"table": tableName,
		})

		ui.Log(ui.SQLLogger, "[%d] Read unique query: \n%s", sessionID, util.SessionLog(sessionID, q))

		rows, err := db.Query(q)
		if err != nil {
			util.ErrorResponse(w, sessionID, err.Error(), http.StatusInternalServerError)

			return http.StatusInternalServerError
		}

		defer rows.Close()

		uniqueColumns := map[string]bool{}
		keys := []string{}

		for rows.Next() {
			var name string

			_ = rows.Scan(&name)
			uniqueColumns[name] = true

			keys = append(keys, name)
		}

		ui.Log(ui.TableLogger, "[%d] Unique columns: %v", sessionID, keys)

		// Determine which columns are nullable.
		q = queryParameters(nullableColumnsQuery, map[string]string{
			"table": tableName,
			"quote": "",
		})

		ui.Log(ui.SQLLogger, "[%d] Read nullable query: %s", sessionID, util.SessionLog(sessionID, q))

		var nrows *sql.Rows

		nrows, err = db.Query(q)
		if err != nil {
			util.ErrorResponse(w, sessionID, err.Error(), http.StatusInternalServerError)

			return http.StatusInternalServerError
		}

		defer nrows.Close()

		nullableColumns := map[string]bool{}
		keys = []string{}

		for nrows.Next() {
			var schemaName, tableName, columnName string

			var nullable bool

			_ = nrows.Scan(&schemaName, &tableName, &columnName, &nullable)

			if nullable {
				nullableColumns[columnName] = true

				keys = append(keys, columnName)
			}
		}

		ui.Log(ui.TableLogger, "[%d] Nullable columns: %v", sessionID, keys)

		// Get standard column names an type info.
		columns, e2 := getColumnInfo(db, user, tableName, sessionID)
		if e2 == nil {
			// Determine which columns are nullable
			for n, column := range columns {
				columns[n].Nullable = nullableColumns[column.Name]
			}

			// Determine which columns are also unique
			for n, column := range columns {
				columns[n].Unique = uniqueColumns[column.Name]
			}

			resp := defs.TableColumnsInfo{
				ServerInfo: util.MakeServerInfo(sessionID),
				Columns:    columns,
				Count:      len(columns),
			}

			w.Header().Add("Content-Type", defs.TableMetadataMediaType)

			b, _ := json.MarshalIndent(resp, "", "  ")
			_, _ = w.Write(b)

			if ui.IsActive(ui.RestLogger) {
				ui.WriteLog(ui.RestLogger, "[%d] Response payload:\n%s", sessionID, util.SessionLog(sessionID, string(b)))
			}

			return http.StatusOK
		}

		if e2 != nil {
			err = errors.NewError(e2)
		}
	}

	msg := fmt.Sprintf("database table metadata error, %s", strings.TrimPrefix(err.Error(), "pq: "))
	status := http.StatusBadRequest

	if strings.Contains(err.Error(), "does not exist") {
		status = http.StatusNotFound
	}

	if err == nil && db == nil {
		msg = unexpectedNilPointerError
		status = http.StatusInternalServerError
	}

	util.ErrorResponse(w, sessionID, msg, status)

	return status
}

func getColumnInfo(db *sql.DB, user string, tableName string, sessionID int) ([]defs.DBColumn, error) {
	columns := make([]defs.DBColumn, 0)
	name, _ := fullName(user, tableName)

	q := queryParameters(tableMetadataQuery, map[string]string{
		"table": name,
	})

	ui.Log(ui.SQLLogger, "[%d] Reading table metadata query: %s", sessionID, q)

	rows, err := db.Query(q)
	if err == nil {
		defer rows.Close()

		names, _ := rows.Columns()
		types, _ := rows.ColumnTypes()

		for i, name := range names {
			// Special case, we synthetically create a defs.RowIDName column
			// and it is always of type "UUID". But we don't return it
			// as a user column name.
			if name == defs.RowIDName {
				continue
			}

			typeInfo := types[i]

			// Start by seeing what Go type it will become. IF that isn't
			// known, then get the underlying database type name instead.
			typeName := typeInfo.ScanType().Name()
			if typeName == "" {
				typeName = typeInfo.DatabaseTypeName()
			}

			size, _ := typeInfo.Length()
			nullable, _ := typeInfo.Nullable()

			columns = append(columns, defs.DBColumn{
				Name:     name,
				Type:     typeName,
				Size:     int(size),
				Nullable: nullable},
			)
		}
	}

	if err != nil {
		return columns, errors.NewError(err)
	}

	return columns, nil
}

// DeleteTable will delete a database table from the user's schema.
func DeleteTable(user string, isAdmin bool, tableName string, sessionID int, w http.ResponseWriter, r *http.Request) int {
	if err := util.AcceptedMediaType(r, []string{}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	// Verify that there are no parameters
	if err := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	tableName, _ = fullName(user, tableName)

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		if !isAdmin && Authorized(sessionID, db, user, tableName, adminOperation) {
			util.ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return http.StatusForbidden
		}

		q := queryParameters(tableDeleteQuery, map[string]string{
			"table": tableName,
		})

		ui.Log(ui.SQLLogger, "[%d] Query: %s", sessionID, q)

		_, err = db.Exec(q)
		if err == nil {
			RemoveTablePermissions(sessionID, db, tableName)
			util.ErrorResponse(w, sessionID, "Table "+tableName+" successfully deleted", http.StatusOK)

			return http.StatusOK
		}
	}

	msg := fmt.Sprintf("database table delete error, %s", strings.TrimPrefix(err.Error(), "pq: "))

	if err == nil && db == nil {
		msg = unexpectedNilPointerError
	}

	status := http.StatusBadRequest
	if strings.Contains(msg, "does not exist") {
		status = http.StatusNotFound
	}

	util.ErrorResponse(w, sessionID, msg, status)

	return status
}

// ListTables will list all the tables for the given user.
func ListTables(user string, isAdmin bool, sessionID int, w http.ResponseWriter, r *http.Request) int {
	if err := util.AcceptedMediaType(r, []string{defs.TablesMediaType}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	if r.Method != http.MethodGet {
		msg := "Unsupported method " + r.Method + " " + r.URL.Path
		util.ErrorResponse(w, sessionID, msg, http.StatusBadRequest)

		return http.StatusBadRequest
	}

	// Verify that the parameters are valid, if given.
	if err := util.ValidateParameters(r.URL, map[string]string{
		defs.StartParameterName:    "int",
		defs.LimitParameterName:    "int",
		defs.UserParameterName:     "string",
		defs.RowCountParameterName: "bool",
	}); err != nil {
		util.ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return http.StatusBadRequest
	}

	// Currently, the default is to include row counts in the listing. You
	// could change this in the future if it proves too inefficient.
	includeRowCounts := true

	v := r.URL.Query()[defs.RowCountParameterName]
	if len(v) == 1 {
		includeRowCounts = data.Bool(v[0])
	}

	db, err := OpenDB(sessionID, user, "")

	if err == nil && db != nil {
		var rows *sql.Rows

		q := strings.ReplaceAll(tablesListQuery, "{{schema}}", user)
		if paging := pagingClauses(r.URL); paging != "" {
			q = q + paging
		}

		ui.Log(ui.ServerLogger, "[%d] attempting to read tables from schema %s", sessionID, user)
		ui.Log(ui.SQLLogger, "[%d] Query: %s", sessionID, q)

		rows, err = db.Query(q)
		if err == nil {
			var name string

			defer rows.Close()

			names := make([]defs.Table, 0)
			count := 0

			for rows.Next() {
				err = rows.Scan(&name)
				if err != nil {
					break
				}

				// Is the user authorized to see this table at all?
				if !isAdmin && Authorized(sessionID, db, user, name, readOperation) {
					continue
				}

				// See how many columns are in this table. Must be a fully-qualfiied name.
				columnQuery := "SELECT * FROM \"" + user + "\".\"" + name + "\" WHERE 1=0"
				ui.Log(ui.SQLLogger, "[%d] Columns metadata query: %s", sessionID, columnQuery)

				tableInfo, err := db.Query(columnQuery)
				if err != nil {
					continue
				}

				defer tableInfo.Close()
				count++

				columns, _ := tableInfo.Columns()
				columnCount := len(columns)

				for _, columnName := range columns {
					if columnName == defs.RowIDName {
						columnCount--

						break
					}
				}

				// Let's also count the rows. This may become too expensive but let's try it.
				rowCount := 0

				if includeRowCounts {
					q := queryParameters(rowCountQuery, map[string]string{
						"schema": user,
						"table":  name,
					})

					ui.Log(ui.SQLLogger, "[%d] Row count query: %s", sessionID, q)

					result, e2 := db.Query(q)
					if e2 != nil {
						util.ErrorResponse(w, sessionID, e2.Error(), http.StatusInternalServerError)

						return http.StatusInternalServerError
					}

					defer result.Close()

					if result.Next() {
						_ = result.Scan(&rowCount)
					}
				}

				// Package up the info for this table to add to the list.
				names = append(names, defs.Table{
					Name:    name,
					Schema:  user,
					Columns: columnCount,
					Rows:    rowCount,
				})
			}

			ui.Log(ui.ServerLogger, "[%d] read %d table names", sessionID, count)

			if err == nil {
				resp := defs.TableInfo{
					ServerInfo: util.MakeServerInfo(sessionID),
					Tables:     names,
					Count:      len(names),
				}

				w.Header().Add("Content-Type", defs.TablesMediaType)

				b, _ := json.MarshalIndent(resp, "", "  ")
				_, _ = w.Write(b)

				if ui.IsActive(ui.RestLogger) {
					ui.WriteLog(ui.RestLogger, "[%d] Response payload:\n%s", sessionID, util.SessionLog(sessionID, string(b)))
				}

				return http.StatusOK
			}
		}
	}

	msg := fmt.Sprintf("Database list error, %v", err)
	if err == nil && db == nil {
		msg = unexpectedNilPointerError
	}

	util.ErrorResponse(w, sessionID, msg, http.StatusBadRequest)

	return http.StatusBadRequest
}

func parameterString(r *http.Request) string {
	m := r.URL.Query()
	result := strings.Builder{}

	for k, v := range m {
		if result.Len() == 0 {
			result.WriteRune('?')
		} else {
			result.WriteRune('&')
		}

		result.WriteString(k)

		if len(v) > 0 {
			result.WriteRune('=')

			for n, value := range v {
				if n > 0 {
					result.WriteRune(',')
				}

				result.WriteString(value)
			}
		}
	}

	return result.String()
}
