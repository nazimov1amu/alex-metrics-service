package utils

import (
	"fmt"
	"time"
)


func LinearRetry(attempts int, interval time.Duration, fn func() error) error {
	var sleepDuration time.Duration = 1 * time.Second

	for i := 1; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		time.Sleep(sleepDuration)
		sleepDuration += interval
	}
	return fmt.Errorf("failed after %d attempts", attempts)
}



