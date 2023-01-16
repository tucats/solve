package rest

import (
	"os"

	"github.com/tucats/ego/app-cli/settings"
	"github.com/tucats/ego/data"
	"github.com/tucats/ego/defs"
	"github.com/tucats/ego/errors"
	"github.com/tucats/ego/symbols"
)

// Do we allow outbound REST calls with invalid/insecure certificates?
var allowInsecure = false

// Verify implements the Verify() rest function. This accepts a boolean value
// and sets the TLS server certificate authentication accordingly. When set to true,
// a connection will not be made if the server's certificate cannot be authenticated.
// This is the default mode for HTTPS connections. During debugging, you may wish to
// turn this off when using self-generated certificates.
func Verify(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.ErrArgumentCount
	}

	this := getThis(s)
	verify := allowInsecure

	if len(args) == 1 {
		verify = data.Bool(args[0])
	}

	this.SetAlways(verifyFieldName, verify)

	return this, nil
}

// Auth implements the Auth() rest function. When present, it accepts a username and
// password as parameters, and sets the rest client to use BasicAuth authentication, where
// the username and password are part of an Authentication header.
func Auth(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	r, err := getClient(s)
	if err != nil {
		return nil, err
	}

	this := getThis(s)

	if len(args) != 2 {
		return nil, errors.ErrArgumentCount
	}

	user := data.String(args[0])
	pass := data.String(args[1])

	r.SetBasicAuth(user, pass)

	return this, nil
}

// Token implements the Token() rest function. When present, it accepts a token string
// and sets the rest client to use Bearer token authentication using this token value.
func Token(s *symbols.SymbolTable, args []interface{}) (interface{}, error) {
	r, err := getClient(s)
	if err != nil {
		return nil, err
	}

	this := getThis(s)

	if len(args) > 1 {
		return nil, errors.ErrArgumentCount
	}

	token := settings.Get(defs.LogonTokenSetting)

	if len(args) > 0 {
		token = data.String(args[0])
	}

	r.SetAuthToken(token)

	return this, nil
}

// Externalized function that sets the "insecure" flag, which turns off
// server validation. This is called from the CLI parsing action when
// "--insecure" is specified as a global option in the Ego command line.
func AllowInsecure(flag bool) {
	allowInsecure = flag

	if flag {
		os.Setenv("EGO_INSECURE_CLIENT", defs.True)
	} else {
		os.Setenv("EGO_INSECURE_CLIENT", "")
	}
}