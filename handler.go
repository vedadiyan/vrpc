package vrpc

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/vedadiyan/vapor"
	"github.com/vedadiyan/vedio"
)

func CreateHandler[T any, R any](serviceName string, method string, pattern vapor.Pattern) (vapor.Pattern, func(vapor.Request) vapor.Response) {
	px, err := CreateProxy(serviceName, reflect.TypeFor[T](), reflect.TypeFor[R]())
	if err != nil {
		return "", nil
	}
	vedio.RegisterFor[func(vapor.Request) vapor.Response, func(vapor.Request) vapor.Response](vedio.WithName(fmt.Sprintf("%s.%s", serviceName, strings.ToLower(method))), vedio.WithGenerator(func() (func(vapor.Request) vapor.Response, error) {
		return func(r vapor.Request) vapor.Response {
			i, err := io.ReadAll(r.Content())
			if err != nil {
				return ToError(serviceName, err)
			}
			req, err := Decode[T](r.Headers().Get("Content-Type"), i)
			if err != nil {
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
		fn, err := vedio.Resolve[func(vapor.Request) vapor.Response](vedio.WithName(fmt.Sprintf("%s.%s", serviceName, strings.ToLower(r.Method()))))
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
