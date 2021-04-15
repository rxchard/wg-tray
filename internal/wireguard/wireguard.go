package wireguard

type WgStat struct {
	Enabled       bool
	Endpoint      string
	BytesSent     uint64
	BytesReceived uint64
}
