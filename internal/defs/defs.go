package defs

type RequestContext struct {
	TraceID string
}

const (
	MethodWebsocket string = "WEBSOCKET"
	PathPrefix      string = "/api/v2"
	EmptyString     string = ""
	StatusPending   int    = 1
)

var ConnectionState = struct {
	Closed     string
	Open       string
	Connecting string
}{
	Closed:     "CLOSED",
	Open:       "OPEN",
	Connecting: "CONNECTING",
}
