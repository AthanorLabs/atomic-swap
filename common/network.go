// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package common

import (
	"fmt"
	"strings"
)

// Environment represents the environment the swap will run in (ie. mainnet, stagenet, or development)
type Environment byte

const (
	// Undefined is a placeholder, do not pass it to functions
	Undefined Environment = iota
	// Mainnet is for real use with mainnet monero and ethereum endpoints
	Mainnet
	// Stagenet is for testing with stagenet monero and gorerli ethereum endpoints
	Stagenet
	// Development is for testing with a local monerod in regtest mode and Ganache simulating ethereum
	Development
)

// String ...
func (env Environment) String() string {
	switch env {
	case Mainnet:
		return "mainnet"
	case Stagenet:
		return "stagenet"
	case Development:
		return "dev"
	}

	return "undefined"
}

// NewEnv converts an environment string into the Environment type
func NewEnv(envStr string) (Environment, error) {
	switch strings.ToLower(envStr) {
	case "mainnet":
		return Mainnet, nil
	case "stagenet":
		return Stagenet, nil
	case "dev":
		return Development, nil
	default:
		return Undefined, fmt.Errorf(`unknown environment %q, expected "mainnet", "stagenet" or "dev"`, envStr)
	}
}
