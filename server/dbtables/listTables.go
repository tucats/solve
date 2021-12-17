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

func ListTables(user string, sessionID int32, w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		msg := "Unsupported method " + r.Method + " " + r.URL.Path
		errorResponse(w, sessionID, msg, http.StatusBadRequest)
		return
	}

	db, err := OpenDB(sessionID, user, "")

	if err == nil && db != nil {
		var rows *sql.Rows

		q := strings.ReplaceAll(tablesQueryString, "{{schema}}", user)

		ui.Debug(ui.ServerLogger, "[%d] attempting to read tables from schema %s", sessionID, user)
		ui.Debug(ui.ServerLogger, "[%d]    with query %s", sessionID, q)

		rows, err = db.Query(q)
		if err == nil {

			names := make([]string, 0)
			var name string
			count := 0

			for rows.Next() {
				err = rows.Scan(&name)
				if err != nil {
					break
				}
				count++
				names = append(names, name)
			}

			ui.Debug(ui.ServerLogger, "[%d] read %d table names", sessionID, count)

			if err == nil {
				resp := defs.TableInfo{
					Tables:       names,
					RestResponse: defs.RestResponse{Status: 200},
				}

				b, _ := json.MarshalIndent(resp, "", "  ")
				w.Write(b)

				return
			}
		}
	}

	msg := fmt.Sprintf("Database list error, %v", err)
	if err == nil && db == nil {
		msg = "Unexpected nil database object pointer"
	}

	errorResponse(w, sessionID, msg, http.StatusBadRequest)
}