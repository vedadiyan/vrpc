package vrpc

import (
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
		method string
	}

	HandlerOption func(*handlerOptions)
)

var (
	handlers sync.Map
)

func CreateHandler[T any, R any](serviceName string, pattern vapor.Pattern, opts ...HandlerOption) (func() (vapor.Pattern, func(vapor.Request) vapor.Response), error) {
	value, _ := handlers.LoadOrStore(serviceName, func(r vapor.Request) vapor.Response {
		fn, err := vedio.Resolve[func(vapor.Request) vapor.Response](vedio.WithName(GetFullServiceName(serviceName, r.Method())))
		if err != nil {
			return ToError(serviceName, err)
		}
		return fn(r)
	})
	px, err := CreateProxy(serviceName, reflect.TypeFor[T](), reflect.TypeFor[R]())
	if err != nil {
		return nil, err
	}
	handlerOptions := &handlerOptions{}
	for _, opt := range opts {
		opt(handlerOptions)
	}
	vedio.RegisterFor[func(vapor.Request) vapor.Response, func(vapor.Request) vapor.Response](vedio.WithName(GetFullServiceName(serviceName, handlerOptions.method)), vedio.WithGenerator(func() (func(vapor.Request) vapor.Response, error) {
		return func(r vapor.Request) vapor.Response {
			i, err := io.ReadAll(r.Content())
			if err != nil {
				return ToError(serviceName, err)
			}
			var req T
			if err := Decode(r.Headers().Get("Content-Type"), i, &req); err != nil {
				return ToError(serviceName, err)
			}
			res, err := px(req)
			if err != nil {
				return ToError(serviceName, err)
			}
			o, err := Encode(r.Headers().Get("Accept"), res)
			if err != nil {
				return ToError(serviceName, err)
			}
			return vapor.NewResponse(MapSuccessCode(serviceName), vapor.WithContent(o))
		}, nil
	}))
	return func() (vapor.Pattern, func(vapor.Request) vapor.Response) {
		return pattern, value.(func(vapor.Request) vapor.Response)
	}, nil
}

func ToError(serviceName string, err error) vapor.Response {
	return vapor.NewResponse(
		MapErrorCode(serviceName, err),
		vapor.WithHeaders(vapor.KeyValue{"Content-Type": {"text/plain"}}),
		vapor.WithContent([]byte(err.Error())),
	)
}

func GetFullServiceName(serviceName string, method string) string {
	if len(method) == 0 {
		return serviceName
	}
	return fmt.Sprintf("%s.%s", strings.ToLower(serviceName), strings.ToLower(method))
}
