package resp

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func OK() Response {
	return Response{Status: StatusOK}
}

func Error(msg string) Response {
	return Response{Status: StatusError, Error: msg}
}
