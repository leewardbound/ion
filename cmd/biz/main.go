package main

import (
	"net/http"
	_ "net/http/pprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	conf "github.com/pion/ion/pkg/conf/biz"
	"github.com/pion/ion/pkg/discovery"
	"github.com/pion/ion/pkg/log"
	"github.com/pion/ion/pkg/node/biz"
	"github.com/pion/ion/pkg/signal"

)

func init() {
	log.Init(conf.Log.Level)
	signal.Init(conf.Signal.Host, conf.Signal.Port, conf.Signal.Cert, conf.Signal.Key, biz.Entry)
}

func close() {
	biz.Close()
}

func main() {
	log.Infof("--- Starting Biz Node ---")

	if conf.Global.Pprof != "" {
		go func() {
			log.Infof("Start pprof on %s", conf.Global.Pprof)
			err := http.ListenAndServe(conf.Global.Pprof, nil)
			if err != nil {
				panic(err)
			}
		}()
	}
	
	if conf.Global.Metrics != "" {
		go func() {
			log.Infof("Serving metrics at %s/metrics", conf.Global.Metrics)
			http.Handle("/metrics", promhttp.Handler())
			err := http.ListenAndServe(conf.Global.Metrics, nil)
			if err != nil {
				panic(err)
			}
		}()
	}

	serviceNode := discovery.NewServiceNode(conf.Etcd.Addrs, conf.Global.Dc)
	serviceNode.RegisterNode("biz", "node-biz", "biz-channel-id")

	rpcID := serviceNode.GetRPCChannel()
	eventID := serviceNode.GetEventChannel()
	biz.Init(conf.Global.Dc, serviceNode.NodeInfo().ID, rpcID, eventID, conf.Nats.URL)

	serviceWatcher := discovery.NewServiceWatcher(conf.Etcd.Addrs, conf.Global.Dc)
	serviceWatcher.WatchServiceNode("islb", biz.WatchServiceNodes)


	defer close()
	select {}
}
