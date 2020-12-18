package logtail

import (
	"os"
)

const TransferTypeConsole = "console"

type ConsoleTransfer struct {
}

func (d *ConsoleTransfer) Trans(serverId string, data []byte) error {
	_, _ = os.Stdout.Write(data)
	_, _ = os.Stdout.Write([]byte{'\n'})
	return nil
}
