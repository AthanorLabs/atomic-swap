// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalStatus(t *testing.T) {
	type S struct {
		Status Status `json:"status"`
	}

	const jsonText = `{
		"status": "XMRLocked"
	}`

	s := new(S)
	err := json.Unmarshal([]byte(jsonText), s)
	require.NoError(t, err)
	require.Equal(t, XMRLocked, s.Status)

	jsonData, err := json.Marshal(s)
	require.NoError(t, err)
	require.JSONEq(t, jsonText, string(jsonData))
}

func TestUnmarshalStatus_fail(t *testing.T) {
	type S struct {
		Status Status `json:"status"`
	}

	const jsonText = `{
		"status": "Garbage"
	}`

	s := new(S)
	err := json.Unmarshal([]byte(jsonText), s)
	require.ErrorContains(t, err, `unknown status "Garbage"`)

	s.Status = 255 // not a valid value
	_, err = json.Marshal(s)
	require.ErrorContains(t, err, `unknown status 255`)
}
