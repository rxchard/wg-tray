package manager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/getlantern/systray"
	"github.com/rxchard/wg-tray/internal/client"
	"github.com/rxchard/wg-tray/internal/icons"
	"io"
	"log"
	"strings"
	"time"
)

var appx context.Context
var c *client.WgClient
var items map[string]*systray.MenuItem

func trayExit() {}

func trayReady() {
	items = map[string]*systray.MenuItem{}

	systray.SetIcon(icons.True())
	systray.SetTitle("")
	systray.AddMenuItem("Wireguard", "Wireguard Connection Manager").Disable()
	systray.AddSeparator()

	items["connect"] = systray.AddMenuItem("Connect", "Connect to the configured Wireguard VPN")

	items["stats_endpoint"] = systray.AddMenuItem("Endpoint: 255.255.255.255", "Wireguard Endpoint")
	items["stats_endpoint"].Disable()

	items["stats_sent"] = systray.AddMenuItem("Sent: none", "Wireguard Transfer")
	items["stats_sent"].Disable()

	items["stats_received"] = systray.AddMenuItem("Received: none", "Wireguard Transfer")
	items["stats_received"].Disable()

	systray.AddSeparator()

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for {
			select {
			case <-appx.Done():
				ticker.Stop()
				log.Println("status ticker exit")
				return
			case <-ticker.C:
				clientGetStatus()
			}
		}
	}()

	clientGetStatus()
	trayLoop()
}

func trayLoop() {
	for {
		select {
		case <-appx.Done():
			return
		case <-items["connect"].ClickedCh:
			// toggle wireguard, i.e. bring interface up/down
			if err := c.Write("toggle"); err == nil {
				c.Stat.Enabled = !c.Stat.Enabled
				clientGetStatus()
			}
		}
	}
}

func trayStatusUpdate() {
	if c.Stat.Enabled {
		items["connect"].SetTitle("Disconnect")
		systray.SetIcon(icons.True())
	} else {
		items["connect"].SetTitle("Connect")
		systray.SetIcon(icons.False())
	}

	items["stats_endpoint"].SetTitle(fmt.Sprintf("Endpoint: %s", c.Stat.Endpoint))

	items["stats_sent"].SetTitle(fmt.Sprintf("Sent: %s",
		datasize.ByteSize(c.Stat.BytesSent).HumanReadable()))

	items["stats_received"].SetTitle(fmt.Sprintf("Received: %s",
		datasize.ByteSize(c.Stat.BytesReceived).HumanReadable()))
}

func clientGetStatus() {
	if err := c.Write("status"); err != nil {
		log.Printf("failed to send status message: %v\n", err)
		return
	}
}

func clientRead(buffer []byte) error {
	n, err := c.Connection.Read(buffer)
	if err != nil {
		return err
	}

	packet := strings.Split(string(buffer[0:n]), ":")
	if len(packet) != 2 {
		return fmt.Errorf("invalid packet: %s", packet)
	}

	out, err := base64.URLEncoding.DecodeString(packet[1])
	if err != nil {
		return err
	}

	if packet[0] == "status" {
		if err = json.Unmarshal(out, &c.Stat); err != nil {
			return err
		}

		trayStatusUpdate()
		return nil
	}

	return fmt.Errorf("unhandled packet: %s", packet)
}

func clientReader(parent context.Context) {
	buffer := make([]byte, 4096)

	for {
		if err := clientRead(buffer); err != nil {
			// did we exit?
			if parent.Err() != nil {
				return
			}

			// did server exit
			if err == io.EOF {
				return
			}

			log.Printf("error while reading: %v\n", err)
		}
	}
}

func Execute(parent context.Context) error {
	var cancel context.CancelFunc
	appx, cancel = context.WithCancel(parent)
	defer cancel()

	var err error
	c, err = client.Execute()
	if err != nil {
		return err
	}

	go func() {
		clientReader(appx)
		log.Println("client read done")
		cancel()
	}()

	go systray.Run(trayReady, trayExit)

	go func() {
		<-appx.Done()
		log.Println("exit loop")
		systray.Quit()
		_ = c.Connection.Close()
	}()

	log.Println("started")

	<-appx.Done()
	return nil
}
