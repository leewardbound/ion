package signal

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/server"
	"github.com/pion/ion/pkg/log"
	"github.com/pion/ion/pkg/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AcceptFunc peer.AcceptFunc
type RejectFunc peer.RejectFunc
type RespondFunc peer.RespondFunc

const (
	errInvalidMethod = "method not found"
	errInvalidData   = "data not found"
	statCycle        = time.Second * 3
)

var (
	bizCall  func(method string, peer *Peer, msg json.RawMessage, accept RespondFunc, reject RejectFunc)
	wsServer *server.WebSocketServer
	rooms    = make(map[proto.RID]*Room)
	roomLock sync.RWMutex

	gaugeRooms = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ion_biz_rooms",
		Help: "The number of rooms open on this biz server",
	})
	gaugePeers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ion_biz_peers",
		Help: "The number of peers connected to this biz server",
	})
)

func Init(host string, port int, cert, key string, bizEntry func(method string, peer *Peer, msg json.RawMessage, accept RespondFunc, reject RejectFunc)) {
	wsServer = server.NewWebSocketServer(in)
	config := server.DefaultConfig()
	config.Host = host
	config.Port = port
	config.CertFile = cert
	config.KeyFile = key
	config.HTMLRoot = "web"
	bizCall = bizEntry
	go wsServer.Bind(config)
	go stat()
}

func stat() {
	t := time.NewTicker(statCycle)
	defer t.Stop()
	for range t.C {
		info := "\n----------------signal-----------------\n"
		print := false
		roomLock.Lock()
		if len(rooms) > 0 {
			print = true
		}
		gaugePeers.Set(0)
		for rid, room := range rooms {
			info += fmt.Sprintf("room: %s\npeers: %d\n", rid, len(room.GetPeers()))
			if len(room.GetPeers()) == 0 {
				delete(rooms, rid)
			}
			gaugePeers.Add(float64(len(room.GetPeers())))
		}
		gaugeRooms.Set(float64(len(rooms)))
		roomLock.Unlock()
		if print {
			log.Infof(info)
		}
	}
}
