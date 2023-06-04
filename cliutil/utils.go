// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package cliutil provides utility functions intended for sharing by the main packages of multiple executables.
package cliutil

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/common"
)

var (
	// Only use this logger in functions called by programs that use formatted logs like swapd (not swapcli)
	log = logging.Logger("cmd")
)

func createAndWriteEthKeyFile(ethPrivKeyFile string, env common.Environment, devXMRMaker, devXMRTaker bool) error {
	var key *ecdsa.PrivateKey
	var err error

	switch {
	case env == common.Development && devXMRMaker:
		key, err = ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRMaker)
	case env == common.Development && devXMRTaker:
		key, err = ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	default:
		key, err = ethcrypto.GenerateKey()
	}
	if err != nil {
		return err
	}

	privKeyStr := hexutil.Encode(ethcrypto.FromECDSA(key))
	privKeyStr = strings.TrimPrefix(privKeyStr, "0x")

	if err := os.WriteFile(ethPrivKeyFile, []byte(privKeyStr), 0600); err != nil {
		return err
	}

	log.Infof("New ETH wallet key generated in %s", ethPrivKeyFile)
	log.Infof("Fund address %s to take an offer",
		ethcrypto.PubkeyToAddress(*(key.Public().(*ecdsa.PublicKey))).Hex())
	return nil
}

// GetEthereumPrivateKey reads or creates and returns an ethereum private key for the given the CLI options.
func GetEthereumPrivateKey(ethPrivKeyFile string, env common.Environment, devXMRMaker, devXMRTaker bool) (
	*ecdsa.PrivateKey,
	error,
) {
	if ethPrivKeyFile == "" {
		panic("missing required parameter ethPrivKeyFile")
	}

	exists, err := common.FileExists(ethPrivKeyFile)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err = createAndWriteEthKeyFile(ethPrivKeyFile, env, devXMRMaker, devXMRTaker); err != nil {
			return nil, err
		}
	}

	fileData, err := os.ReadFile(filepath.Clean(ethPrivKeyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to read eth-privkey file: %w", err)
	}
	ethPrivKeyHex := strings.TrimSpace(string(fileData))
	privkey, err := ethcrypto.HexToECDSA(ethPrivKeyHex)
	if err != nil {
		return nil, err
	}

	if exists {
		log.Infof("Using ETH wallet key located in %s", ethPrivKeyFile)
		log.Infof("ETH address: %s", ethcrypto.PubkeyToAddress(*(privkey.Public().(*ecdsa.PublicKey))).Hex())
	}

	return privkey, nil
}

// GetVersion returns our version string for an executable
func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown-version"
	}

	commitHash := ""
	sourcesModified := false

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commitHash = setting.Value
		case "vcs.modified":
			if setting.Value == "true" {
				sourcesModified = true
			}
		}
	}

	var version strings.Builder

	// The first part a go.mod style version if using "go install" directly with
	// github. If installing from locally checked out source, the string will be
	// "(devel)".
	version.WriteString(strings.Replace(info.Main.Version, "(devel)", "dev-", 1))

	// The commit hash will be present if installing from locally checked out
	// sources, or empty if installing directly from the repo's github URL.
	if commitHash != "" {
		// 7 bytes is what "git rev-parse --short HEAD" returns
		version.WriteString(fmt.Sprintf("%.7s", commitHash))
		if sourcesModified {
			version.WriteString("-dirty")
		}
	}

	version.WriteByte('-')
	version.WriteString(info.GoVersion)

	return version.String()
}

// ReadUnsignedDecimalFlag reads a string flag and parses it into an *apd.Decimal.
func ReadUnsignedDecimalFlag(ctx *cli.Context, flagName string) (*apd.Decimal, error) {
	s := ctx.String(flagName)
	if s == "" {
		return nil, fmt.Errorf("flag --%s cannot be empty", flagName)
	}
	bf, _, err := new(apd.Decimal).SetString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid value %q for flag --%s", s, flagName)
	}
	if bf.IsZero() {
		return nil, fmt.Errorf("value of flag --%s cannot be zero", flagName)
	}
	if bf.Negative {
		return nil, fmt.Errorf("value of flag --%s cannot be negative", flagName)
	}

	return bf, nil
}

// ReadETHAddress reads a string flag and parses to an ethereum Address type
func ReadETHAddress(ctx *cli.Context, flagName string) (*ethcommon.Address, error) {
	s := ctx.String(flagName)
	if s == "" {
		return nil, fmt.Errorf("flag --%s cannot be empty", flagName)
	}

	ok := ethcommon.IsHexAddress(s)
	if !ok {
		return nil, fmt.Errorf("invalid ETH address: %q", s)
	}

	to := ethcommon.HexToAddress(s)

	return &to, nil
}

// ExpandBootnodes expands the boot nodes passed on the command line that
// can be specified individually with multiple flags, but can also contain
// multiple boot nodes passed to single flag separated by commas.
func ExpandBootnodes(nodesCLI []string) []string {
	var nodes []string // nodes from all flag values combined
	for _, flagVal := range nodesCLI {
		splitNodes := strings.Split(flagVal, ",")
		for _, n := range splitNodes {
			n = strings.TrimSpace(n)
			// Handle the empty string to not use default bootnodes. Doing it here after
			// the split has the arguably positive side effect of skipping empty entries.
			if len(n) > 0 {
				nodes = append(nodes, strings.TrimSpace(n))
			}
		}
	}
	return nodes
}
