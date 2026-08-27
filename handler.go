package vrpc

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/vedadiyan/vapor"
	"github.com/vedadiyan/vedio"
)

type (
	handlerOptions struct {
		serviceName string
		method      string
	}

	HandlerOption func(*handlerOptions)
)

func CreateHandler[T any, R any](serviceName string, pattern vapor.Pattern, opts ...HandlerOption) (vapor.Pattern, func(vapor.Request) vapor.Response) {
	px, err := CreateProxy(serviceName, reflect.TypeFor[T](), reflect.TypeFor[R]())
	if err != nil {
		return "", nil
	}
	handlerOptions := &handlerOptions{
		serviceName: strings.ToLower(serviceName),
	}
	for _, opt := range opts {
		opt(handlerOptions)
	}
	vedio.RegisterFor[func(vapor.Request) vapor.Response, func(vapor.Request) vapor.Response](vedio.WithName(handlerOptions.GetFullServiceName()), vedio.WithGenerator(func() (func(vapor.Request) vapor.Response, error) {
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
	return pattern, func(r vapor.Request) vapor.Response {
		fn, err := vedio.Resolve[func(vapor.Request) vapor.Response](vedio.WithName(handlerOptions.GetFullServiceName()))
		if err != nil {
			return ToError(serviceName, err)
		}
		return fn(r)
	}
}

func ToError(serviceName string, err error) vapor.Response {
	return vapor.NewResponse(
		MapErrorCode(serviceName, err),
		vapor.WithHeaders(vapor.KeyValue{"Content-Type": {"text/plain"}}),
		vapor.WithContent([]byte(err.Error())),
	)
}

func (h *handlerOptions) GetFullServiceName() string {
	if len(h.method) == 0 {
		return h.serviceName
	}
	return fmt.Sprintf("%s.%s", h.serviceName, h.method)
}
