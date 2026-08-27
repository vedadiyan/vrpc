package vrpc

type (
	Error        string
	ErrorPattern interface {
		Pattern() string
	}
)

func (e Error) Error() string {
	return string(e)
}
