package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/tucats/ego/app-cli/cli"
	"github.com/tucats/ego/app-cli/tables"
	"github.com/tucats/ego/app-cli/ui"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/runtime"
)

// SetCacheSize is the administrative command that sets the server's cache size for
// storing previously-compiled service handlers. If you specify a smaller number
// that the current cache size, the next attempt to load a new service into the cache
// will result in discarding the oldest cache entries until the cache is the correct
// size. You must be an admin user with a valid token to perform this command.
func SetCacheSize(c *cli.Context) *errors.EgoError {
	if c.GetParameterCount() == 0 {
		return errors.New(errors.ErrCacheSizeNotSpecified)
	}

	size, err := strconv.Atoi(c.GetParameter(0))
	if !errors.Nil(err) {
		return errors.New(err)
	}

	cacheStatus := defs.CacheResponse{
		Limit: size,
	}

	err = runtime.Exchange("/admin/caches", "POST", &cacheStatus, &cacheStatus)
	if !errors.Nil(err) {
		return errors.New(err)
	}

	switch ui.OutputFormat {
	case ui.JSONFormat:
		b, _ := json.Marshal(cacheStatus)

		fmt.Println(string(b))

	case ui.JSONIndentedFormat:
		b, _ := json.MarshalIndent(cacheStatus, ui.JSONIndentPrefix, ui.JSONIndentSpacer)

		fmt.Println(string(b))

	case ui.TextFormat:
		if cacheStatus.Status != http.StatusOK {
			if cacheStatus.Status == http.StatusForbidden {
				return errors.New(errors.ErrNoPrivilegeForOperation)
			}

			return errors.NewMessage(cacheStatus.Message)
		}

		ui.Say("Server cache size updated")
	}

	return nil
}

// FlushServerCaches is the administrative command that directs the server to
// discard any cached compilation units for service code. Subsequent service
// requests require that the service code be reloaded from disk. This is often
// used when making changes to a service, to quickly force the server to pick up
// the changes. You must be an admin user with a valid token to perform this command.
func FlushServerCaches(c *cli.Context) *errors.EgoError {
	cacheStatus := defs.CacheResponse{}

	err := runtime.Exchange("/admin/caches", "DELETE", nil, &cacheStatus)
	if !errors.Nil(err) {
		return err
	}

	switch ui.OutputFormat {
	case ui.JSONIndentedFormat:
		b, _ := json.MarshalIndent(cacheStatus, ui.JSONIndentPrefix, ui.JSONIndentSpacer)

		fmt.Println(string(b))

	case ui.JSONFormat:
		b, _ := json.Marshal(cacheStatus)

		fmt.Println(string(b))

	case ui.TextFormat:
		if cacheStatus.Status != http.StatusOK {
			if cacheStatus.Status == http.StatusForbidden {
				return errors.New(errors.ErrNoPrivilegeForOperation)
			}

			return errors.NewMessage(cacheStatus.Message)
		}

		ui.Say("Server cache emptied")
	}

	return nil
}

// ListServerCahces is the administrative command that displays the information about
// the server's cache of previously-compiled service programs. The current and maximum
// size of the cache, and the endpoints that are cached are listed. You must be an
// admin user with a valid token to perform this command.
func ListServerCaches(c *cli.Context) *errors.EgoError {
	cacheStatus := defs.CacheResponse{}

	err := runtime.Exchange("/admin/caches", "GET", nil, &cacheStatus)
	if !errors.Nil(err) {
		return err
	}

	if cacheStatus.Status != http.StatusOK {
		return errors.New(errors.ErrHTTP).Context(cacheStatus.Status)
	}

	switch ui.OutputFormat {
	case ui.JSONIndentedFormat:
		b, _ := json.MarshalIndent(cacheStatus, ui.JSONIndentPrefix, ui.JSONIndentSpacer)
		fmt.Println(string(b))

	case ui.JSONFormat:
		b, _ := json.Marshal(cacheStatus)
		fmt.Println(string(b))

	case ui.TextFormat:
		if cacheStatus.Status != http.StatusOK {
			if cacheStatus.Status == http.StatusForbidden {
				return errors.New(errors.ErrNoPrivilegeForOperation)
			}

			return errors.NewMessage(cacheStatus.Message)
		}

		fmt.Printf("Server cache status (%d/%d) items\n", cacheStatus.Count, cacheStatus.Limit)

		if cacheStatus.Count > 0 {
			fmt.Printf("\n")

			t, _ := tables.New([]string{"Endpoint", "Count", "Last Used"})

			for _, v := range cacheStatus.Items {
				_ = t.AddRowItems(v.Name, v.Count, v.LastUsed)
			}

			t.Print("text")
		}
	}

	return nil
}