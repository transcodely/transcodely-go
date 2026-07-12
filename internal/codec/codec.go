// Package codec provides the Transcodely JSON wire codec.
//
// This file is a verbatim copy of github.com/transcodely/api/internal/connect/codec.go
// at commit eca70c69a494b2d9cc75b79aea5580324a99a77b. Do NOT edit here — edit upstream
// in the api repo and resync. CI verifies the two files match.
package codec

import (
	"encoding/json"
	"strings"
	"unicode"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoJSONCodec is a JSON codec that uses snake_case field names
// and simplified lowercase enum values.
// This provides a REST-like JSON format matching our database conventions.
//
// Example transformations:
//   - field names: "basePath" → "base_path"
//   - enums: "JOB_STATUS_PENDING" → "pending"
//   - enums: "ORIGIN_PROVIDER_GCS" → "gcs"
//
// The codec uses proto reflection to automatically handle all enums,
// including future additions, without any hardcoded mappings.
type ProtoJSONCodec struct {
	marshalOptions   protojson.MarshalOptions
	unmarshalOptions protojson.UnmarshalOptions
}

// NewProtoJSONCodec creates a new JSON codec with snake_case field names
// and simplified enum values.
func NewProtoJSONCodec() *ProtoJSONCodec {
	return &ProtoJSONCodec{
		marshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true, // Use snake_case field names
			EmitUnpopulated: true, // Always emit all fields for consistent API responses
		},
		unmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true, // Be lenient with unknown fields
		},
	}
}

// Name returns the codec name.
func (c *ProtoJSONCodec) Name() string {
	return "json"
}

// Marshal serializes a protobuf message to JSON with snake_case field names
// and simplified enum values.
func (c *ProtoJSONCodec) Marshal(msg any) ([]byte, error) {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return nil, errNotProtoMessage
	}

	// Marshal with protojson to get proper field names
	data, err := c.marshalOptions.Marshal(protoMsg)
	if err != nil {
		return nil, err
	}

	// Parse to map, simplify enums, re-serialize
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	transformEnumsInMap(m, protoMsg.ProtoReflect().Descriptor(), simplifyEnumValue)

	return json.Marshal(m)
}

// Unmarshal deserializes JSON into a protobuf message.
// Accepts both simplified enum values (e.g., "pending") and full names (e.g., "JOB_STATUS_PENDING").
func (c *ProtoJSONCodec) Unmarshal(data []byte, msg any) error {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return errNotProtoMessage
	}

	// Parse to map
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// Expand simplified enum values using the message descriptor (context-aware)
	transformEnumsInMap(m, protoMsg.ProtoReflect().Descriptor(), expandEnumValue)

	// Re-serialize and let protojson do the final parsing
	expanded, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.unmarshalOptions.Unmarshal(expanded, protoMsg)
}

// enumTransformFunc transforms an enum value string using the enum descriptor.
type enumTransformFunc func(value string, enumDesc protoreflect.EnumDescriptor) string

// transformEnumsInMap recursively transforms enum values in a JSON map
// using the message descriptor for context-aware transformation.
func transformEnumsInMap(m map[string]any, desc protoreflect.MessageDescriptor, transform enumTransformFunc) {
	for key, value := range m {
		field := findField(desc, key)
		if field == nil {
			continue
		}

		switch {
		case field.IsMap():
			// Handle map fields with enum values
			if mapVal, ok := value.(map[string]any); ok {
				if field.MapValue().Kind() == protoreflect.EnumKind {
					for k, v := range mapVal {
						if strVal, ok := v.(string); ok {
							mapVal[k] = transform(strVal, field.MapValue().Enum())
						}
					}
				} else if field.MapValue().Kind() == protoreflect.MessageKind {
					for _, v := range mapVal {
						if subMap, ok := v.(map[string]any); ok {
							transformEnumsInMap(subMap, field.MapValue().Message(), transform)
						}
					}
				}
			}

		case field.IsList():
			// Handle repeated fields (arrays)
			if list, ok := value.([]any); ok {
				for i, item := range list {
					if field.Kind() == protoreflect.EnumKind {
						if strVal, ok := item.(string); ok {
							list[i] = transform(strVal, field.Enum())
						}
					} else if field.Kind() == protoreflect.MessageKind {
						if subMap, ok := item.(map[string]any); ok {
							transformEnumsInMap(subMap, field.Message(), transform)
						}
					}
				}
			}

		case field.Kind() == protoreflect.EnumKind:
			// Handle scalar enum fields
			if strVal, ok := value.(string); ok {
				m[key] = transform(strVal, field.Enum())
			}

		case field.Kind() == protoreflect.MessageKind:
			// Handle nested message fields
			if subMap, ok := value.(map[string]any); ok {
				transformEnumsInMap(subMap, field.Message(), transform)
			}
		}
	}
}

// findField finds a field descriptor by JSON name or proto name.
func findField(desc protoreflect.MessageDescriptor, key string) protoreflect.FieldDescriptor {
	fields := desc.Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		if f.JSONName() == key || string(f.Name()) == key {
			return f
		}
	}
	return nil
}

// expandEnumValue converts a simplified value to the full protobuf enum name.
// "pending" + JobStatus descriptor → "JOB_STATUS_PENDING"
// "pending" + OutputStatus descriptor → "OUTPUT_STATUS_PENDING"
func expandEnumValue(value string, enumDesc protoreflect.EnumDescriptor) string {
	values := enumDesc.Values()

	// If already a valid full enum name, return as-is
	if values.ByName(protoreflect.Name(value)) != nil {
		return value
	}

	// Derive the expected prefix from the enum type name
	// "JobStatus" → "JOB_STATUS_"
	prefix := enumPrefix(enumDesc)

	// Handle multi-word values like "awaiting_confirmation" → "AWAITING_CONFIRMATION"
	upperValue := strings.ToUpper(value)

	// Construct the full enum name
	fullName := prefix + upperValue

	// Validate it exists in this enum type
	if values.ByName(protoreflect.Name(fullName)) != nil {
		return fullName
	}

	// Return original value - let protojson handle/reject invalid values
	return value
}

// simplifyEnumValue converts a full protobuf enum name to a simplified lowercase value.
// "JOB_STATUS_PENDING" → "pending"
// "VIDEO_CODEC_H264" → "h264"
func simplifyEnumValue(value string, enumDesc protoreflect.EnumDescriptor) string {
	prefix := enumPrefix(enumDesc)
	if strings.HasPrefix(value, prefix) {
		return strings.ToLower(strings.TrimPrefix(value, prefix))
	}
	return strings.ToLower(value)
}

// enumPrefix derives the enum value prefix from the enum type name.
// "JobStatus" → "JOB_STATUS_"
// "VideoCodec" → "VIDEO_CODEC_"
// "Resolution" → "RESOLUTION_"
func enumPrefix(enumDesc protoreflect.EnumDescriptor) string {
	name := string(enumDesc.Name())
	return toScreamingSnake(name) + "_"
}

// toScreamingSnake converts CamelCase to SCREAMING_SNAKE_CASE.
// "JobStatus" → "JOB_STATUS"
// "VideoCodec" → "VIDEO_CODEC"
// "HLSSegmentFormat" → "HLS_SEGMENT_FORMAT"
func toScreamingSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			// Check if previous char was lowercase or next char is lowercase
			// This handles "APIKey" → "API_KEY" correctly
			prev := rune(s[i-1])
			if unicode.IsLower(prev) {
				result.WriteRune('_')
			} else if i+1 < len(s) && unicode.IsLower(rune(s[i+1])) {
				result.WriteRune('_')
			}
		}
		result.WriteRune(unicode.ToUpper(r))
	}
	return result.String()
}

// errNotProtoMessage is returned when a non-proto message is passed to the codec.
var errNotProtoMessage = &codecError{msg: "message is not a proto.Message"}

type codecError struct {
	msg string
}

func (e *codecError) Error() string {
	return e.msg
}
