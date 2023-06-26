package infra

import "errors"

var (
	gIsCnEmpty = CnEmptyError{}
	gIsFull    = tsdbEfull{}
	gIsEmpty   = tsdbEEmpty{}
	gIsEof     = tsdEEof{}
	gIsBEmpty  = blotIsEmpty{}
	gIsBExist  = blotIsExist{}
)

func isError(err error, target error) bool {
	if err == nil {
		return false
	}
	return !isTargetError(err, target)
}

func isTargetError(err error, target error) bool {
	if err == target {
		return true
	}
	return errors.Is(err, target)
}

type tsdbEfull struct {
}

type tsdbEEmpty struct {
}

type tsdEEof struct {
}

type CnEmptyError struct {
}

type blotIsEmpty struct {
}

type blotIsExist struct {
}

// tsdbFullError
func (tsfe tsdbEfull) Error() string {
	return "Is full"
}

// tsdbEEmpty
func (tsfee tsdbEEmpty) Error() string {
	return "Is Empty"
}

func (tsdeof tsdEEof) Error() string {
	return "Eof"
}

func (ee CnEmptyError) Error() string {
	return "CnEmptyError"
}

func (ee blotIsEmpty) Error() string {
	return "CnEmptyError"
}

func (ee blotIsExist) Error() string {
	return "CnEmptyError"
}
