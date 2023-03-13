package utils

import (
	"encoding/json"
	"github.com/liaoyudong2/GateServer/zlog"
	"os"
)

const LoggerPath = "log/GateSrv"

// GateSSLConfig 网关的ssl配置
type GateSSLConfig struct {
	Open   bool
	Cert   string
	PKey   string
	Passwd string
}

// GateConfig 网管配置
type GateConfig struct {
	UseSSL         GateSSLConfig
	BindClientPort int
	BindSrvAddr    string
}

type GameConfig struct {
	BindSrvAddr string
}

type ServerConfig struct {
	GateSrv  GateConfig
	ServerId int
}

var GlobalConfig *ServerConfig

func (g *ServerConfig) Reload() {
	data, err := os.ReadFile("config/SrvCfg.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, GlobalConfig)
	if err != nil {
		panic(err)
	}
}

func init() {
	zlog.SetLogPath(LoggerPath)
	GlobalConfig = &ServerConfig{}
	GlobalConfig.Reload()
}
