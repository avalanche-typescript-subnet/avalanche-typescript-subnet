package runtime

import "fmt"

type CallbackFunc func([]byte) ([]byte, error)

func EmptyCallbackFunc([]byte) ([]byte, error) {
	return nil, fmt.Errorf("empty callback")
}

func _emptyCallbackFunc(int32, int32) int64 { return 0 }

var callback CallbackFunc = EmptyCallbackFunc

func SetCallbackFunc(cb CallbackFunc) {
	if cb == nil {
		callback = EmptyCallbackFunc
	} else {
		callback = cb
	}
}
