package flag

import (
	"errors"
	"strconv"
)

var ErrHelp = errors.New("flag:help requested")

var errRange = errors.New("value out of range")

var errParse = errors.New("parse error")

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errParse
	}
	return err
}
