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
		method      string
		errCodes    map[string]int
		successCode int
	}

	HandlerOption func(*handlerOptions)
)

var (
	handlers sync.Map
)

const (
	HeaderContentType = "Content-Type"
	HeaderAccept      = "Accept"

	MimeTypeTextPlain = "text/plain"
)

func CreateHandler[T any, R any](serviceName string, pattern vapor.Pattern, opts ...HandlerOption) (func() (vapor.Pattern, func(vapor.Request) vapor.Response), error) {
	px, err := CreateProxy(serviceName, reflect.TypeFor[T](), reflect.TypeFor[R]())
	if err != nil {
		return nil, err
	}

	handlerOptions := &handlerOptions{
		successCode: 200,
	}
	for _, opt := range opts {
		opt(handlerOptions)
	}

	generator := func() (func(vapor.Request) vapor.Response, error) {
		return func(r vapor.Request) vapor.Response {
			i, err := io.ReadAll(r.Content())
			if err != nil {
				return ToError(handlerOptions.errCodes, err)
			}
			var req T
			if err := Decode(r.Headers().Get(HeaderContentType), i, &req); err != nil {
				return ToError(handlerOptions.errCodes, err)
			}
			res, err := px(req)
			if err != nil {
				return ToError(handlerOptions.errCodes, err)
			}
			o, err := Encode(r.Headers().Get(HeaderAccept), res)
			if err != nil {
				return ToError(handlerOptions.errCodes, err)
			}
			return vapor.NewResponse(handlerOptions.successCode, vapor.WithContent(o))
		}, nil
	}

	if err := vedio.Register[func(vapor.Request) vapor.Response](vedio.WithName(GetFullServiceName(serviceName, handlerOptions.method)), vedio.WithGenerator(generator)); err != nil {
		return nil, err
	}

	value, _ := handlers.LoadOrStore(strings.ToLower(serviceName), func(r vapor.Request) vapor.Response {
		fn, err := vedio.Resolve[func(vapor.Request) vapor.Response](vedio.WithName(GetFullServiceName(serviceName, r.Method())))
		if err != nil {
			return vapor.NewResponse(
				500,
				vapor.WithHeaders(vapor.KeyValue{HeaderContentType: {MimeTypeTextPlain}}),
				vapor.WithContent([]byte(err.Error())),
			)
		}
		return fn(r)
	})

	return func() (vapor.Pattern, func(vapor.Request) vapor.Response) {
		return pattern, value.(func(vapor.Request) vapor.Response)
	}, nil
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
