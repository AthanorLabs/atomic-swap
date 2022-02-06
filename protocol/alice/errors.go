package alice

import (
	"errors"
)

var (
	errSwapAborted           = errors.New("swap cancelled early, but before any locking happened")
	errUnexpectedMessageType = errors.New("unexpected message type")
	errMissingKeys           = errors.New("did not receive Bob's public spend or private view key")
	errMissingAddress        = errors.New("did not receive Bob's address")
	errNoClaimLogsFound      = errors.New("no Claimed logs found")
)
