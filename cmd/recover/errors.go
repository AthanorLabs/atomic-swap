package main

import (
	"errors"
)

var (
	errNoSecretsProvided               = errors.New("must also provide one of --alice-secret or --bob-secret")
	errNoAliceSecretOrContractProvided = errors.New("must also provide one of --alice-secret or --contract-addr")
	errNoBobSecretOrContractProvided   = errors.New("must also provide one of --contract-addr or --bob-secret")
)
