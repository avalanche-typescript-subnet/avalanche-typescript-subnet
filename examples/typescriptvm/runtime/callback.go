package runtime

import "fmt"

type CallbackFunc func([]byte) ([]byte, error)

func EmptyCallbackFunc([]byte) ([]byte, error) {
	return nil, fmt.Errorf("empty callback")
}

func (exec *JavyExec) SetCallbackFunc(cb CallbackFunc) *JavyExec {
	if cb == nil {
		exec.callback = EmptyCallbackFunc
	} else {
		exec.callback = cb
	}
	return exec
}
