package vjson

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
)

// Documentation on the validator package is here:
// https://pkg.go.dev/github.com/go-playground/validator/v10
var validate = validator.New()

// MarshalStruct adds additional validation on top of json.Marshal. Input type
// should be a struct pointer.
func MarshalStruct(v any) ([]byte, error) {
	if err := validate.Struct(v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// MarshalIndentStruct adds additional validation on top of json.MarshalIndent.
// Input type should be a struct pointer.
func MarshalIndentStruct(v any, prefix, indent string) ([]byte, error) {
	if err := validate.Struct(v); err != nil {
		return nil, err
	}
	return json.MarshalIndent(v, prefix, indent)
}

// UnmarshalStruct adds additional validation on top of json.Unmarshal. Target
// object be a struct pointer.
func UnmarshalStruct(jsonData []byte, v any) error {
	if err := json.Unmarshal(jsonData, v); err != nil {
		return err
	}
	return validate.Struct(v)
}
