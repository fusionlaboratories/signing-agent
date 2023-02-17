package defs

const AuthHeader = "x-api-zkp"

type RequestContext struct {
	TraceID string
}

const (
	MethodWebsocket = "WEBSOCKET"
	PathPrefix      = "/api/v1"
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
