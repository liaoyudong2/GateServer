package main

import (
	"github.com/liaoyudong2/GateServer/net"
	"github.com/liaoyudong2/GateServer/utils"
	"github.com/liaoyudong2/GateServer/zlog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	zlog.SetLogConsole()

	go net.Instance().StartService(utils.GlobalConfig.GateSrv.BindClientPort + utils.GlobalConfig.ServerId)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		break
	}
	zlog.Warn("GateSrv Shutdown...")
	net.Instance().StopService()
	zlog.Warn("GateSrv Shutdown...Done")
}
