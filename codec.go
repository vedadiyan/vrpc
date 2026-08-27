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

const (
	MimeTypeApplicationJson     = "application/json"
	MimeTypeApplicationProtobuf = "application/protobuf"
	MimeTypeTextToon            = "text/toon"
)

func Decode(contentType string, in []byte, v any) error {
	switch GetMimeType(contentType) {
	case MimeTypeApplicationJson:
		{
			if err := json.Unmarshal(in, v); err != nil {
				return err
			}
			return nil
		}
	case MimeTypeApplicationProtobuf:
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
	case MimeTypeTextToon:
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
	switch GetMimeType(contentType) {
	case MimeTypeApplicationJson:
		{
			return json.Marshal(in)
		}
	case MimeTypeApplicationProtobuf:
		{
			r, ok := any(in).(codecs.Reflected)
			if !ok {
				return nil, ErrIncompatibleCodec
			}
			return protolizer.StaticCodec().Marshal(r)
		}
	case MimeTypeTextToon:
		{
			return toon.Marshal(in)
		}
	default:
		{
			return nil, ErrContentTypeUnspecified
		}
	}
}

func GetMimeType(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
}
