// Package vjson or "validated JSON" provides additional validation, configured
// via annotations, on structures as they are Marshaled or Unmarshalled to/from
// JSON data.
package vjson

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
)

type SomeStruct struct {
	Decimal *apd.Decimal `json:"decimal" validate:"required"`
	Hex     string       `json:"hex,omitempty" validate:"omitempty,hexadecimal"`
}

func TestMarshalStruct(t *testing.T) {
	s := &SomeStruct{Decimal: apd.New(11, -1)}

	data, err := MarshalStruct(s)
	require.NoError(t, err)
	require.Equal(t, `{"decimal":"1.1"}`, string(data))

	data, err = MarshalIndentStruct(s, "", "  ")
	require.NoError(t, err)
	require.Equal(t, "{\n  \"decimal\": \"1.1\"\n}", string(data))
}

func TestMarshalStruct_notValid(t *testing.T) {
	s := &SomeStruct{Decimal: nil}
	errMsg := `'SomeStruct.Decimal' Error:Field validation for 'Decimal' failed on the 'required' tag`

	_, err := MarshalStruct(s)
	require.Error(t, err)
	require.ErrorContains(t, err, errMsg)

	_, err = MarshalIndentStruct(s, "", "  ")
	require.Error(t, err)
	require.ErrorContains(t, err, errMsg)
}

func TestUnmarshalStruct(t *testing.T) {
	var s = new(SomeStruct)
	err := UnmarshalStruct([]byte(`{"decimal":"0","hex":"0x12ab"}`), s)
	require.NoError(t, err)
}

func TestUnmarshalStruct_notValid(t *testing.T) {
	var s = new(SomeStruct)
	err := UnmarshalStruct([]byte(`{"decimal":"0","hex":"xyz"}`), s)
	errMsg := `Key: 'SomeStruct.Hex' Error:Field validation for 'Hex' failed on the 'hexadecimal' tag`
	require.ErrorContains(t, err, errMsg)
}
