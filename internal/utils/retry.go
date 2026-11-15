// Package utils содержит повтор запросов и запуск миграций базы.
package utils

import (
	"fmt"
	"time"
)

// LinearRetry вызывает fn до attempts раз.
// Пауза перед повтором начинается с одной секунды и каждый раз увеличивается на interval.
// При успехе возвращается nil, иначе ошибка исчерпания попыток.
func LinearRetry(attempts int, interval time.Duration, fn func() error) error {
	sleepDuration := 1 * time.Second

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
