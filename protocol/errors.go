package protocol

import (
	"errors"
)

var (
	errInvalidSecp256k1Key = errors.New("secp256k1 public key resulting from proof verification does not match key sent")
)
