package vrpc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/toon-format/toon-go"
	"github.com/vedadiyan/protolizer"
	"github.com/vedadiyan/protolizer/codecs"
)

func Decode[T any](contentType string, in []byte) (*T, error) {
	var zero T
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "application/json":
		{
			if err := json.Unmarshal(in, &zero); err != nil {
				return nil, err
			}
			return &zero, nil
		}
	case "application/protobuf":
		{
			r, ok := any(&zero).(codecs.Reflected)
			if !ok {
				return nil, fmt.Errorf("incompatible protobuffer codec detected")
			}
			if err := protolizer.StaticCodec().Unmarshal(in, r); err != nil {
				return nil, err
			}
			return &zero, nil
		}
	case "text/toon":
		{
			if err := toon.Unmarshal(in, &zero); err != nil {
				return nil, err
			}
			return &zero, nil
		}
	default:
		{
			return nil, fmt.Errorf("content type unspecified")
		}
	}
}

func Encode[T any](contentType string, in T) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "application/json":
		{
			return json.Marshal(in)
		}
	case "application/protobuf":
		{
			r, ok := any(&in).(codecs.Reflected)
			if !ok {
				return nil, fmt.Errorf("incompatible protobuffer codec detected")
			}
			return protolizer.StaticCodec().Marshal(r)
		}
	case "text/toon":
		{
			return toon.Marshal(in)
		}
	default:
		{
			return nil, fmt.Errorf("content type unspecified")
		}
	}
}
