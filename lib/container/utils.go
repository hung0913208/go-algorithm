package container

// #cgo CFLAGS: -g -Wall
// #include <unistd.h>
import "C"

import (
	"errors"
)

func alarm(seconds uint) (remaining uint, err error) {
	remaining = uint(C.alarm(C.uint(seconds)))
	if remaining == 0 {
		err = errors.New("alarm fail")
	}
	return
}
