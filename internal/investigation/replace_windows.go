//go:build windows

package investigation

import (
	"time"

	"golang.org/x/sys/windows"
)

func replaceFile(source, target string) error {
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(2 * time.Second)
	for {
		err = windows.MoveFileEx(
			from,
			to,
			windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH,
		)
		if err == nil {
			return nil
		}
		if err != windows.ERROR_ACCESS_DENIED &&
			err != windows.ERROR_SHARING_VIOLATION &&
			err != windows.ERROR_LOCK_VIOLATION {
			return err
		}
		if time.Now().After(deadline) {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
}
