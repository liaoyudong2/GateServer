package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liaoyudong2/GateServer/net/iface"
	"github.com/liaoyudong2/GateServer/zlog"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

// BridgeService
// @Description: 桥服务
type BridgeService struct {
	listener    net.Listener      // 监听对象
	lock        sync.RWMutex      // 读写锁
	exitChan    chan bool         // 退出信号
	sessionIter atomic.Uint32     // 会话自增ID
	maxSession  int               // 最大会话数量
	sessionMgr  iface.ISessionMgr // 连接管理
}

const DefaultMaxSession = 1024

var gateService = &BridgeService{
	exitChan:   make(chan bool, 1),
	maxSession: DefaultMaxSession,
	sessionMgr: NewSessionMgr(DefaultMaxSession),
}

func Ins() *BridgeService {
	return gateService
}

func (gs *BridgeService) StartService(port int) {
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
				_ = conn.Close()
				return
			}
			gs.sessionIter.Add(1)
			gs.sessionMgr.AddSession(NewSession(gs.sessionIter.Load(), conn, gs.sessionMgr))
		})

		zlog.Infof("BridgeService startup, listen at %v", addr)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
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

func (gs *BridgeService) SendMessageToSession(sessionId uint32, msg iface.IMessage) {
	session := gs.sessionMgr.GetSession(sessionId)
	if session == nil {
		zlog.Errorf("session undefined, session id: %d", sessionId)
	} else {
		session.SendMessage(msg)
	}
}

func (gs *BridgeService) RawBufferToSession(sessionId uint32, buf []byte) {
	session := gs.sessionMgr.GetSession(sessionId)
	if session == nil {
		zlog.Errorf("session undefined, session id: %d", sessionId)
	} else {
		session.RawBuffer(buf)
	}
}
