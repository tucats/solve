package dbtables

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/datatypes"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/util"
)

const UnexpectedNilPointerError = "Unexpected nil database object pointer"

// TableCreate creates a new table based on the JSON payload, which must be an array of DBColumn objects, defining
// the characteristics of each column in the table. If the table name is the special name "@sql" the payload instead
// is assumed to be a JSON-encoded string containing arbitrary SQL to exectue. Only an admin user can use the "@sql"
// table name.
func TableCreate(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	var err error

	// Verify that there are no parameters
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	db, err := OpenDB(sessionID, user, "")

	if err == nil && db != nil {
		// Special case; if the table name is @SQL then the payload is processed as a simple
		// SQL  statement and not a table creation.
		if strings.EqualFold(tableName, sqlPseudoTable) {
			if !isAdmin {
				ErrorResponse(w, sessionID, "No privilege for direct SQL execution", http.StatusForbidden)

				return
			}

			var statementText string

			err := json.NewDecoder(r.Body).Decode(&statementText)
			if err == nil {
				if strings.HasPrefix(strings.TrimSpace(strings.ToLower(statementText)), "select ") {
					ui.Debug(ui.TableLogger, "[%d] SQL query: %s", sessionID, statementText)

					err = readRowData(db, statementText, sessionID, w)
					if err != nil {
						ErrorResponse(w, sessionID, "Error reading SQL query; "+err.Error(), http.StatusInternalServerError)

						return
					}
				} else {
					var rows sql.Result

					ui.Debug(ui.TableLogger, "[%d] SQL execute: %s", sessionID, statementText)

					rows, err = db.Exec(statementText)
					if err == nil {
						count, _ := rows.RowsAffected()
						reply := defs.DBRowCount{Count: int(count)}

						b, _ := json.MarshalIndent(reply, "", "  ")
						_, _ = w.Write(b)

						return
					}
				}
			}

			if err != nil {
				ErrorResponse(w, sessionID, "Error in SQL execute; "+err.Error(), http.StatusInternalServerError)
			}

			return
		}

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

		for _, column := range data {
			if column.Name == "" {
				ErrorResponse(w, sessionID, "Missing or empty column name", http.StatusBadRequest)

				return
			}

			if column.Type == "" {
				ErrorResponse(w, sessionID, "Missing or empty type name", http.StatusBadRequest)

				return
			}

			if !keywordMatch(column.Type, defs.TableColumnTypeNames...) {
				ErrorResponse(w, sessionID, "Invalid type name: "+column.Type, http.StatusBadRequest)

				return
			}
		}

		q := formCreateQuery(r.URL, user, isAdmin, data, sessionID, w)
		if q == "" {
			return
		}

		if !createSchemaIfNeeded(w, sessionID, db, user, tableName) {
			return
		}

		ui.Debug(ui.TableLogger, "[%d] Create table with query: %s", sessionID, q)

		counts, err := db.Exec(q)
		if err == nil {
			rows, _ := counts.RowsAffected()
			result := defs.DBRowCount{
				Count: int(rows),
			}

			tableName, _ = fullName(user, tableName)
			CreateTablePermissions(sessionID, db, user, tableName, readOperation, deleteOperation, updateOperation)

			b, _ := json.MarshalIndent(result, "", "  ")
			_, _ = w.Write(b)

			ui.Debug(ui.ServerLogger, "[%d] table created", sessionID)

			return
		}

		ui.Debug(ui.ServerLogger, "[%d] Error creating table, %v", sessionID, err)
		ErrorResponse(w, sessionID, err.Error(), http.StatusBadRequest)

		return
	}

	ui.Debug(ui.TableLogger, "[%d] Error inserting into table, %v", sessionID, err)
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

	q := queryParameters(createSchemaQuery, map[string]string{
		"schema": schema,
	})

	result, err := db.Exec(q)
	if err != nil {
		ErrorResponse(w, sessionID, "Error creating schema; "+err.Error(), http.StatusInternalServerError)

		return false
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		ui.Debug(ui.TableLogger, "[%d] Created schema %s", sessionID, schema)
	}

	return true
}

// ReadTable reads the metadata for a given table, and returns it as an array
// of column names and types.
func ReadTable(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	// Verify that there are no parameters
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		// Special case; if the table name is @permissions then the payload is processed as request
		// to read all the permissions data
		if strings.EqualFold(tableName, permissionsPseudoTable) {
			if !isAdmin {
				ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

				return
			}

			ReadAllPermissions(db, sessionID, w, r)

			return
		}

		tableName, _ = fullName(user, tableName)

		if !isAdmin && Authorized(sessionID, nil, user, tableName, readOperation) {
			ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return
		}

		// Determine which columns must be unique
		q := queryParameters(uniqueColumnsQuery, map[string]string{
			"table": tableName,
			"quote": "",
		})

		rows, err := db.Query(q)
		if err != nil {
			ErrorResponse(w, sessionID, err.Error(), http.StatusInternalServerError)

			return
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

		ui.Debug(ui.TableLogger, "[%d] Unique columns: %v", sessionID, keys)

		// Determine which columns are nullable.
		q = queryParameters(nullableColumnsQuery, map[string]string{
			"table": tableName,
			"quote": "",
		})

		nrows, err := db.Query(q)
		if err != nil {
			ErrorResponse(w, sessionID, err.Error(), http.StatusInternalServerError)

			return
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

		ui.Debug(ui.TableLogger, "[%d] Nullable columns: %v", sessionID, keys)

		// Get standard column names an type info.
		columns, e2 := getColumnInfo(db, user, tableName, sessionID)
		if errors.Nil(e2) {
			// Determine which columns are nullable
			for n, column := range columns {
				columns[n].Nullable = nullableColumns[column.Name]
			}

			// Determine which columns are also unique
			for n, column := range columns {
				columns[n].Unique = uniqueColumns[column.Name]
			}

			resp := defs.TableColumnsInfo{
				Columns: columns,
				Count:   len(columns),
			}

			b, _ := json.MarshalIndent(resp, "", "  ")
			_, _ = w.Write(b)

			return
		}

		err = e2
	}

	msg := fmt.Sprintf("database table metadata error, %v", err)
	status := http.StatusBadRequest

	if strings.Contains(err.Error(), "does not exist") {
		status = http.StatusNotFound
	}

	if err == nil && db == nil {
		msg = UnexpectedNilPointerError
		status = http.StatusInternalServerError
	}

	ErrorResponse(w, sessionID, msg, status)
}

func getColumnInfo(db *sql.DB, user string, tableName string, sessionID int32) ([]defs.DBColumn, *errors.EgoError) {
	columns := make([]defs.DBColumn, 0)
	name, _ := fullName(user, tableName)

	q := queryParameters(tableMetadataQuery, map[string]string{
		"table": name,
	})

	ui.Debug(ui.TableLogger, "[%d] Reading table metadata with query %s", sessionID, q)

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
		return columns, errors.New(err)
	}

	return columns, nil
}

//DeleteTable will delete a database table from the user's schema.
func DeleteTable(user string, isAdmin bool, tableName string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	// Verify that there are no parameters
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.UserParameterName: "string",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	tableName, _ = fullName(user, tableName)

	db, err := OpenDB(sessionID, user, "")
	if err == nil && db != nil {
		if !isAdmin && Authorized(sessionID, nil, user, tableName, adminOperation) {
			ErrorResponse(w, sessionID, "User does not have read permission", http.StatusForbidden)

			return
		}

		q := queryParameters(tableDeleteQuery, map[string]string{
			"table": tableName,
		})

		ui.Debug(ui.ServerLogger, "[%d] attempting to delete table %s", sessionID, tableName)
		ui.Debug(ui.TableLogger, "[%d]    with query %s", sessionID, q)

		_, err = db.Exec(q)
		if err == nil {
			RemoveTablePermissions(sessionID, db, tableName)
			ErrorResponse(w, sessionID, "Table "+tableName+" successfully deleted", http.StatusOK)

			return
		}
	}

	msg := fmt.Sprintf("database table delete error, %v", err)
	if err == nil && db == nil {
		msg = UnexpectedNilPointerError
	}

	status := http.StatusBadRequest
	if strings.Contains(msg, "does not exist") {
		status = http.StatusNotFound
	}

	ErrorResponse(w, sessionID, msg, status)
}

// ListTables will list all the tables for the given user.
func ListTables(user string, isAdmin bool, sessionID int32, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		msg := "Unsupported method " + r.Method + " " + r.URL.Path
		ErrorResponse(w, sessionID, msg, http.StatusBadRequest)

		return
	}

	// Verify that the parameters are valid, if given.
	if invalid := util.ValidateParameters(r.URL, map[string]string{
		defs.StartParameterName:    "int",
		defs.LimitParameterName:    "int",
		defs.UserParameterName:     "string",
		defs.RowCountParameterName: "bool",
	}); !errors.Nil(invalid) {
		ErrorResponse(w, sessionID, invalid.Error(), http.StatusBadRequest)

		return
	}

	// Currently, the default is to include row counts in the listing. You
	// could change this in the future if it proves too inefficient.
	includeRowCounts := true

	v := r.URL.Query()[defs.RowCountParameterName]
	if len(v) == 1 {
		includeRowCounts = datatypes.GetBool(v[0])
	}

	db, err := OpenDB(sessionID, user, "")

	if err == nil && db != nil {
		var rows *sql.Rows

		q := strings.ReplaceAll(tablesListQuery, "{{schema}}", user)
		if paging := pagingClauses(r.URL); paging != "" {
			q = q + paging
		}

		ui.Debug(ui.ServerLogger, "[%d] attempting to read tables from schema %s", sessionID, user)
		ui.Debug(ui.TableLogger, "[%d]    with query %s", sessionID, q)

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
				ui.Debug(ui.TableLogger, "[%d] Reading columns metadata with query %s", sessionID, columnQuery)

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

					ui.Debug(ui.TableLogger, "[%d] Reading row count with query %s", sessionID, q)

					result, e2 := db.Query(q)
					if !errors.Nil(e2) {
						ErrorResponse(w, sessionID, e2.Error(), http.StatusInternalServerError)

						return
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

			ui.Debug(ui.ServerLogger, "[%d] read %d table names", sessionID, count)

			if err == nil {
				resp := defs.TableInfo{
					Tables: names,
					Count:  len(names),
				}

				b, _ := json.MarshalIndent(resp, "", "  ")
				_, _ = w.Write(b)

				return
			}
		}
	}

	msg := fmt.Sprintf("Database list error, %v", err)
	if err == nil && db == nil {
		msg = UnexpectedNilPointerError
	}

	ErrorResponse(w, sessionID, msg, http.StatusBadRequest)
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
