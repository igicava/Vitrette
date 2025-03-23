package resilience

import (
	"errors"
	"fmt"
	"math"
	"time"
)

// Retry function retries the operation function maxRetries times with an exponential backoff (pattern: retry pattern)
func Retry(operation func() error, maxRetries int, baseDelay time.Duration) error {
	var err error
	for n := 1; n <= maxRetries; n++ {
		err = operation()
		if err != nil {
			fmt.Printf("Retrying attempt %d...", n)
			time.Sleep(baseDelay * time.Duration(math.Pow(2, float64(n))))
			continue
		}
		return nil
	}
	return err
}

// Timeout function executes the operation function with a timeout (pattern: timeout pattern)
func Timeout(operation func() error, timeout time.Duration) error {
	errChan := make(chan error, 1)

	timer := time.AfterFunc(timeout, func() {
		fmt.Println("Timeout")
		errChan <- errors.New("operation timed out")
	})
	defer timer.Stop()

	go func() {
		errChan <- operation()
	}()

	select {
	case err := <-errChan:
		return err
	}
}

// ProcessWithDLQ function for dead letter queue pattern
func ProcessWithDLQ(messages []string,
	operation func(msg string) error,
	dlq *[]string,
) error {
	var err error
	for _, message := range messages {
		err = operation(message)
		if err != nil {
			*dlq = append(*dlq, message)
		}
	}
	return err
}
