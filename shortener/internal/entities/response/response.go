package response

const (
	StatusOK  = "OK"
	StatusErr = "Error"
)

// Response модель ответа сервера
type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Result any    `json:"result,omitempty"`
}

func Error(msg string) Response {
	return Response{
		Status: StatusErr,
		Error:  msg,
	}
}

func OK(result any) Response {
	return Response{
		Status: StatusOK,
		Result: result,
	}
}
