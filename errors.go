package lemur

// Error is the base error type for api errors.
type Error string

func (e Error) Error() string {
	return string(e)
}

// Errors returned by the api.
const (
	ErrTemplateDir = Error("lemur: error with supplied template directory")
)
