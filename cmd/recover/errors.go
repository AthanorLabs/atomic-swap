package main

import (
	"errors"
)

var (
	errMustSpecifyXMRMakerOrTaker = errors.New("must specify --xmrmaker or --xmrtaker")
	errMustProvideInfoFile        = errors.New("must provide path to swap info file with --infofile")
)
