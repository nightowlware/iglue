// +build linux

package iglue

import (
	"fmt"
	"syscall"
)

func createFifo(path string) (string, error) {
	err := syscall.Mknod(path, syscall.S_IFIFO|0666, 0)

	if err != nil {
		return "FIFO_FAIL", fmt.Errorf("Could not create fifo: %s : %s", path, err.Error())
	}

	return path, nil
}
