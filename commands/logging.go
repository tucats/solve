package commands

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tucats/ego/app-cli/cli"
	"github.com/tucats/ego/app-cli/persistence"
	"github.com/tucats/ego/app-cli/tables"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/runtime"
)

func Logging(c *cli.Context) *errors.EgoError {
	addr := persistence.Get(defs.ApplicationServerSetting)
	if addr == "" {
		addr = persistence.Get(defs.LogonServerSetting)
		if addr == "" {
			addr = "localhost"
		}
	}

	addr = strings.TrimPrefix(addr, "https://")
	addr = strings.TrimPrefix(addr, "http://")

	if c.GetParameterCount() > 0 {
		addr = c.GetParameter(0)
		// If it's valid but has no port number, and --port was not
		// given on the command line, assume the default port 8080
		if u, err := url.Parse("https://" + addr); err == nil {
			if u.Port() == "" && !c.WasFound("port") {
				addr = addr + ":8080"
			}
		}

		if c.WasFound("port") {
			port, _ := c.GetInteger("port")
			addr = fmt.Sprintf("%s:%d", addr, port)
		}
	}

	_, err := getProtocol(addr)
	if !errors.Nil(err) {
		return err
	}

	loggers := defs.LoggingItem{Loggers: map[string]bool{}}
	response := defs.LoggingResponse{}

	if c.WasFound("enable") || c.WasFound("disable") {
		if c.WasFound("enable") {
			loggerNames, _ := c.GetStringList("enable")

			for _, loggerName := range loggerNames {
				logger := ui.Logger(loggerName)
				if logger < 0 {
					return errors.New(errors.ErrInvalidLoggerName).Context(strings.ToUpper(loggerName))
				}

				if logger == ui.ServerLogger {
					continue
				}

				loggers.Loggers[loggerName] = true
			}
		}

		if c.WasFound("disable") {
			loggerNames, _ := c.GetStringList("disable")

			for _, loggerName := range loggerNames {
				logger := ui.Logger(loggerName)
				if logger < 0 || logger == ui.ServerLogger {
					return errors.New(errors.ErrInvalidLoggerName).Context(strings.ToUpper(loggerName))
				}

				if _, ok := loggers.Loggers[loggerName]; ok {
					return errors.New(errors.ErrLoggerConflict).Context(loggerName)
				}

				loggers.Loggers[loggerName] = false
			}
		}

		// Send the update, get a reply
		err := runtime.Exchange("/admin/loggers/", "POST", &loggers, &response)
		if !errors.Nil(err) {
			return err
		}
	} else {
		// No changes, just ask for status
		err := runtime.Exchange("/admin/loggers/", "GET", nil, &response)
		if !errors.Nil(err) {
			return err
		}
	}

	// Formulate the output.

	switch ui.OutputFormat {
	case "text":
		t, _ := tables.New([]string{"Logger", "Active"})

		for k, v := range response.Loggers {
			_ = t.AddRowItems(k, v)
		}

		_ = t.SortRows(0, true)
		t.Print(ui.OutputFormat)

	case "json":
		b, _ := json.Marshal(response.Loggers)
		ui.Say(string(b))

	case "indented":
		b, _ := json.MarshalIndent(response.Loggers, "", "   ")
		ui.Say(string(b))
	}

	return nil
}

func getProtocol(addr string) (string, *errors.EgoError) {
	protocol := ""

	resp := struct {
		Pid     int    `json:"pid"`
		Session string `json:"session"`
		Since   string `json:"since"`
	}{}

	if _, err := url.Parse("https://" + addr); err != nil {
		return protocol, errors.New(err)
	}

	protocol = "https"

	persistence.SetDefault(defs.ApplicationServerSetting, "https://"+addr)

	err := runtime.Exchange("/services/up/", "GET", nil, &resp)
	if !errors.Nil(err) {
		protocol = "http"

		persistence.SetDefault(defs.ApplicationServerSetting, "http://"+addr)

		err := runtime.Exchange("/services/up/", "GET", nil, &resp)
		if !errors.Nil(err) {
			return "", nil
		}
	}

	return protocol, nil
}
