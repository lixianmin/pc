package lane

import "errors"

// Errors that can occur during message handling.
var (
	ErrReplyShouldBeNotNull           = errors.New("reply must not be null")
	ErrReceivedMsgSmallerThanExpected = errors.New("received less data than expected, EOF")
	ErrReceivedMsgBiggerThanExpected  = errors.New("received more data than expected")

	KeyContext = "console.context"
)
