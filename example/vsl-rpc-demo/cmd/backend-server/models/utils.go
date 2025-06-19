package models

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/panjf2000/ants/v2"
)

func SubmitWithRetry[T any](pool *ants.PoolWithFuncGeneric[T], payload T, maxRetries int, delay time.Duration) error {
	var err error
	for i := range maxRetries {
		err = pool.Invoke(payload)
		if err == nil {
			return nil // Success
		}
		log.Printf("Pool full, retrying (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return err // Return last error if all retries failed
}

func extractPredictedClass(s string) (string, error) {
	re := regexp.MustCompile(`'predicted_class':\s*'([^']*)'`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		return matches[1], nil
	}
	return "", fmt.Errorf("predicted_class not found")
}
