package vrpc

import (
	"fmt"
	"reflect"

	"github.com/vedadiyan/automapper"
	"github.com/vedadiyan/vedio"
)

func CreateProxy(name string, inputType reflect.Type, outputType reflect.Type) (func(in any) (any, error), error) {
	service, err := vedio.ResolveAnnonymous(vedio.WithName(name))
	if err != nil {
		return nil, err
	}
	typ := reflect.TypeOf(service)
	if typ.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected function but got %s", typ.Kind())
	}
	if typ.NumIn() != 1 {
		return nil, fmt.Errorf("expected exactly 1 input arg but got %d", typ.NumIn())
	}
	if typ.NumOut() != 2 {
		return nil, fmt.Errorf("expected exactly 2 output args but got %d", typ.NumOut())
	}
	if !typ.Out(1).Implements(reflect.TypeFor[error]()) {
		return nil, fmt.Errorf("expected the last output arg to be error but got %s", typ.Out(1).Name())
	}

	val := reflect.ValueOf(service)

	inTarget := automapper.TypeFrom(typ.In(0))
	outSource := automapper.TypeFrom(typ.Out(0))
	inCodec := automapper.CreateCodec(automapper.TypeFrom(inputType), inTarget)
	outCodec := automapper.CreateCodec(outSource, automapper.TypeFrom(outputType))

	return func(in any) (any, error) {
		inValue := automapper.New(inTarget.ConcreteType())
		if err := inCodec(automapper.ValueOf(in).ConcreteValue(), inValue.ConcreteValue()); err != nil {
			return nil, err
		}
		inVal := inValue.Reference(inTarget.PointerCount())
		out := val.Call([]reflect.Value{inVal.Value})
		if !out[1].IsNil() {
			return nil, out[1].Interface().(error)
		}
		outValue := automapper.New(automapper.TypeFrom(outputType).ConcreteType())
		if err := outCodec(automapper.ValueFrom(out[0]).ConcreteValue(), outValue.ConcreteValue()); err != nil {
			return nil, err
		}
		return outValue.Interface(), nil
	}, nil
}
