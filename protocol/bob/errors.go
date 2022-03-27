package bob

import (
	"errors"
)

var (
	errUnexpectedMessageType = errors.New("unexpected message type")
	errMissingKeys           = errors.New("did not receive Alice's public spend or view key")
	errMissingAddress        = errors.New("got empty contract address")
	errNoRefundLogsFound     = errors.New("no refund logs found")
	errPastClaimTime         = errors.New("past t1, can no longer claim")
)
