package vrpc

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/vedadiyan/vapor"
	"github.com/vedadiyan/vedio"
)

type (
	handlerOptions struct {
		method         string
		errCodes       map[string]int
		successCode    int
		streamRequest  bool
		streamResponse bool
	}

	HandlerOption func(*handlerOptions)

	RegisterFunc func() (vapor.Pattern, func(vapor.Request) vapor.Response)

	HandlerFunc = func(vapor.Request) vapor.Response
)

var (
	handlers sync.Map
)

const (
	HeaderContentType = "Content-Type"
	HeaderAccept      = "Accept"

	MimeTypeTextPlain = "text/plain"
)

func WithMethod(method string) HandlerOption {
	return func(ho *handlerOptions) {
		ho.method = method
	}
}

func WithSuccessCode(successCode int) HandlerOption {
	return func(ho *handlerOptions) {
		ho.successCode = successCode
	}
}

func WithStreamRequest() HandlerOption {
	return func(ho *handlerOptions) {
		ho.streamRequest = true
	}
}

func WithStreamResponse() HandlerOption {
	return func(ho *handlerOptions) {
		ho.streamResponse = true
	}
}

func RegisterHandler[T any, R any](serviceName string, pattern vapor.Pattern, opts ...HandlerOption) error {
	handlerOptions := &handlerOptions{
		successCode: 200,
	}
	for _, opt := range opts {
		opt(handlerOptions)
	}

	px, err := CreateProxy(serviceName, reflect.TypeFor[T](), reflect.TypeFor[R]())
	if err != nil {
		return err
	}

	fullServiceNameStatic := GetFullServiceName(serviceName, handlerOptions.method)

	generator := createGenerator[T, R](px, handlerOptions)

	if err := vedio.Register[HandlerFunc](
		vedio.WithName(fullServiceNameStatic), vedio.WithGenerator(generator)); err != nil {
		return err
	}

	registerFunc := func() (vapor.Pattern, func(r vapor.Request) vapor.Response) {
		return pattern, func(r vapor.Request) vapor.Response {
			fullServiceName := GetFullServiceName(serviceName, handlerOptions.method)
			fn, err := vedio.Resolve[HandlerFunc](vedio.WithName(fullServiceName))
			if err != nil {
				return vapor.NewResponse(
					500,
					vapor.WithHeaders(vapor.KeyValue{HeaderContentType: {MimeTypeTextPlain}}),
					vapor.WithContent([]byte(err.Error())),
				)
			}
			return fn(r)
		}
	}

	_, _ = handlers.LoadOrStore(strings.ToLower(serviceName), registerFunc)

	return nil
}

func createGenerator[T any, R any](px func(in any) (any, error), handleroptions *handlerOptions) func() (func(vapor.Request) vapor.Response, error) {
	return func() (func(r vapor.Request) vapor.Response, error) {
		fn := func(r vapor.Request) vapor.Response {
			i, err := io.ReadAll(r.Content())
			if err != nil {
				return ToError(handleroptions.errCodes, err)
			}
			var req T
			if err := Decode(r.Headers().Get(HeaderContentType), i, &req); err != nil {
				return ToError(handleroptions.errCodes, err)
			}
			res, err := px(req)
			if err != nil {
				return ToError(handleroptions.errCodes, err)
			}
			o, err := Encode(r.Headers().Get(HeaderAccept), res)
			if err != nil {
				return ToError(handleroptions.errCodes, err)
			}
			return vapor.NewResponse(handleroptions.successCode, vapor.WithContent(o))
		}
		return fn, nil
	}

}

func ToError(errCodes map[string]int, err error) vapor.Response {
	errCode := 500
	var errPattern ErrorPattern
	if errors.As(err, &errPattern) {
		if value, ok := errCodes[errPattern.Pattern()]; ok {
			errCode = value
		}
	}
	return vapor.NewResponse(
		errCode,
		vapor.WithHeaders(vapor.KeyValue{HeaderContentType: {MimeTypeTextPlain}}),
		vapor.WithContent([]byte(err.Error())),
	)
}

func GetFullServiceName(serviceName string, method string) string {
	serviceName = strings.ToLower(serviceName)
	method = strings.ToLower(method)
	if len(method) == 0 {
		return serviceName
	}
	return fmt.Sprintf("%s.%s", serviceName, method)
}
