package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liaoyudong2/GateServer/net/iface"
	"github.com/liaoyudong2/GateServer/zlog"
	"net"
	"net/http"
	"sync"
)

// BridgeService
// @Description: 桥服务
type BridgeService struct {
	ipVersion   string            // 协议版本
	listener    net.Listener      // 监听对象
	lock        sync.RWMutex      // 读写锁
	exitChan    chan bool         // 退出信号
	sessionIter uint32            // 会话自增ID
	maxSession  int               // 最大会话数量
	sessionMgr  iface.ISessionMgr // 连接管理
}

const DefaultMaxSession = 1024

var gateService = &BridgeService{
	ipVersion:   "tcp4",
	exitChan:    make(chan bool, 1),
	sessionIter: 0,
	maxSession:  DefaultMaxSession,
	sessionMgr:  NewSessionMgr(DefaultMaxSession),
}

func Instance() *BridgeService {
	return gateService
}

func (gs *BridgeService) StartService(port int, version string) {
	if gs.listener != nil {
		zlog.Error("BridgeService is already startup")
	} else {
		addr := fmt.Sprintf("0.0.0.0:%d", port)

		var websocketUpgrade = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			conn, err := websocketUpgrade.Upgrade(writer, request, nil)
			if err != nil {
				zlog.Errorf("accept tcp error: %v", err)
				return
			}
			if gs.sessionMgr.GetSessionCount() >= gs.maxSession {
				zlog.Error("session count overflow, limit count is %d", gs.maxSession)
				return
			}
			gs.sessionIter++
			gs.sessionMgr.AddSession(NewSession(gs.sessionIter, conn))
		})
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
		zlog.Infof("BridgeService startup, listen at %v", addr)

		select {
		case <-gs.exitChan:
			if err := gs.listener.Close(); err != nil {
				zlog.Error("BridgeService is shutdown error ", err)
			} else {
				zlog.Infof("BridgeService shutdown")
			}
			gs.listener = nil
			// 会话关闭
			gs.sessionMgr.CleanSession()
			// 释放结束信号
			gs.exitChan <- true
			break
		}
	}
}

func (gs *BridgeService) StopService() {
	if gs.listener != nil {
		gs.exitChan <- true
		// 等待结束信号
		select {
		case <-gs.exitChan:
			break
		}
		close(gs.exitChan)
	}
}

func (gs *BridgeService) SetMaxSession(num int) {
	gs.maxSession = num
}

func (gs *BridgeService) GetSessionMgr() iface.ISessionMgr {
	return gs.sessionMgr
}
