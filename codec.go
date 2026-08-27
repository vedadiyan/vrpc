package vrpc

import (
	"encoding/json"
	"strings"

	"github.com/toon-format/toon-go"
	"github.com/vedadiyan/protolizer"
	"github.com/vedadiyan/protolizer/codecs"
)

const (
	ErrIncompatibleCodec      Error = "incompatible protobuffer codec detected"
	ErrContentTypeUnspecified Error = "content type unspecified"
)

func Decode(contentType string, in []byte, v any) error {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "application/json":
		{
			if err := json.Unmarshal(in, v); err != nil {
				return err
			}
			return nil
		}
	case "application/protobuf":
		{
			r, ok := any(v).(codecs.Reflected)
			if !ok {
				return ErrIncompatibleCodec
			}
			if err := protolizer.StaticCodec().Unmarshal(in, r); err != nil {
				return err
			}
			return nil
		}
	case "text/toon":
		{
			if err := toon.Unmarshal(in, v); err != nil {
				return err
			}
			return nil
		}
	default:
		{
			return ErrContentTypeUnspecified
		}
	}
}

func Encode(contentType string, in any) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])) {
	case "application/json":
		{
			return json.Marshal(in)
		}
	case "application/protobuf":
		{
			r, ok := any(in).(codecs.Reflected)
			if !ok {
				return nil, ErrIncompatibleCodec
			}
			return protolizer.StaticCodec().Marshal(r)
		}
	case "text/toon":
		{
			return toon.Marshal(in)
		}
	default:
		{
			return nil, ErrContentTypeUnspecified
		}
	}
}
