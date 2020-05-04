package chordio

import "github.com/pkg/errors"

var (
	errInvalidBindFormat = errors.New("bind must be of the form '[IP]:<port>'")
	errInvalidBindIP     = errors.New("the requested bind IP address is invalid")
	errUnableToGetBindIP = errors.New("unable to get a bind IP")
	errNodeIDConflict    = errors.New("conflict node id")
)
