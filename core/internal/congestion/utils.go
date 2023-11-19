package congestion

import (
	"github.com/apernet/quic-go"
	"github.com/ppoonk/shy/core/internal/congestion/bbr"
	"github.com/ppoonk/shy/core/internal/congestion/brutal"
)

func UseBBR(conn quic.Connection) {
	conn.SetCongestionControl(bbr.NewBbrSender(
		bbr.DefaultClock{},
		bbr.GetInitialPacketSize(conn.RemoteAddr()),
	))
}

func UseBrutal(conn quic.Connection, tx uint64) {
	conn.SetCongestionControl(brutal.NewBrutalSender(tx))
}
