package selection

import "errors"

var ErrUnknownMethod = errors.New("unknown selection method")

var ErrTooManyAttempts = errors.New("failed to select quotes after multiple attempts")
