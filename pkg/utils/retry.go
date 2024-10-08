package utils

import (
	"time"
)

func Retry(attempts int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(sleep)
		sleep *= 2 // Exponential increase in waiting time
	}
	return err
}

//Use in wallet.go:
// err := utils.Retry(3, time.Second, func() error {
//     return tonClient.SendTransaction(privateKey, toAddress, amount)
// })
