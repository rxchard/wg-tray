package client

import (
	"fmt"
	"github.com/rxchard/wg-tray/internal/wireguard"
	"net"
	"time"
)

const socketAddr = "/var/run/wg-tray-daemon.sock"

type WgClient struct {
	Connection net.Conn
	Stat       wireguard.WgStat
}

func (w *WgClient) Write(message ...interface{}) error {
	_, err := fmt.Fprint(w.Connection, message...)
	time.Sleep(1 * time.Millisecond)
	return err
}

func Execute() (*WgClient, error) {
	con, err := net.Dial("unix", socketAddr)
	if err != nil {
		return nil, err
	}

	stats := wireguard.WgStat{}

	return &WgClient{
		Connection: con,
		Stat:       stats,
	}, nil
}
