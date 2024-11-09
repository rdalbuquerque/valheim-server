package godinerrors

type ErrorCode string

const (
	MissingColumnError ErrorCode = "missingColumnError"
)

type ReadError struct {
	Code    ErrorCode
	Message string
}

func (re ReadError) Error() string {
	return re.Message
}
