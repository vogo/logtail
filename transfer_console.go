package logtail

import (
	"os"
)

const TransferTypeConsole = "console"

type ConsoleTransfer struct {
}

func (d *ConsoleTransfer) Trans(serverID string, data ...[]byte) error {
	for _, b := range data {
		_, _ = os.Stdout.Write(b)

		n := len(b)
		if n > 0 && b[n-1] != '\n' {
			_, _ = os.Stdout.Write([]byte{'\n'})
		}
	}

	return nil
}

func (d *ConsoleTransfer) start(*Router) error { return nil }
