package server

// This string is pre-pended to any service handler read from the services/ directory
// during service initialization. This is where any extra code needed to support a
// handler can be written as native Ego code.
var handlerProlog = `
type Request struct {
    Method         string
    Url            string 
    Endpoint       string
    Media          string
    Headers        map[string][]string 
    Parameters     map[string][]string
    Authentication string
    Username       string
    Body           string
}

type Response struct {
	Status         int
	Buffer         string
}

func (r *Response) WriteStatus(status int) {
	@status status
}
func (r *Response) Write(msg string) {
	@response msg
}

func (r *Response) WriteJSON( i interface{}) {
	msg := json.Marshal(i)
	r.Write(msg)
}

func NewResponse() Response {
	r := Response{
		Status:   200,
	}

	return r
}

func NewRequest() Request {
    r := Request{
        Url:             _url,
        Endpoint:        _path_endpoint,
        Headers:         _headers,
        Parameters:      _parms,
        Method:          _method,
        Body:            _body,
    }

    if _json {
        r.Media = "json"
    } else {
        r.Media = "text"
    }

    if _authenticated {
        if _token == "" {
            r.Authentication = "basic"
            r.Username = _user
        } else {
            r.Authentication = "token"
        }
    } else {
        r.Authentication = "none"
    }
    return r
}`

// This code is appended after the service. At a minimum, it should contain
// the @handler directive which runs the handler function by name.
var handlerEpilog = `
@handler handler
`
