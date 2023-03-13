package net

import (
	"github.com/gorilla/websocket"
	"github.com/liaoyudong2/GateServer/net/iface"
	"github.com/liaoyudong2/GateServer/zlog"
	"sync"
)

type Session struct {
	sessionId uint32          // 会话ID
	conn      *websocket.Conn // 连接对象
	closed    bool            // 是否已关闭
	lock      sync.RWMutex    // 读写锁
	netStream iface.IStream   // 解析
}

func NewSession(sessionId uint32, conn *websocket.Conn) iface.ISession {
	session := &Session{
		sessionId: sessionId,
		conn:      conn,
		closed:    false,
		netStream: NewStream(),
	}
	// 启动读
	go session.startReader()
	// 启动写
	go session.startWriter()
	return session
}

func (s *Session) GetSessionId() uint32 {
	return s.sessionId
}

func (s *Session) Close() {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.closed == true {
		return
	}
	_ = s.conn.Close()
	s.closed = true
	zlog.Info("session close, session id is ", s.sessionId)
}

func (s *Session) startReader() {
	zlog.Info("session reader start: ", s.sessionId)
	defer s.Close()

	for {
		msgType, data, err := s.conn.ReadMessage()
		if err != nil {
			zlog.Error("session read error: ", err)
			return
		}
		if msgType != websocket.BinaryMessage {
			zlog.Error("session read msg type error")
			return
		}
		readLen := 0
		dataLen := len(data)
		for {
			nread, message, err := s.netStream.Unmarshal(data)
			if err != nil {
				zlog.Error("session net stream unmarshal error: ", err)
				return
			}
			readLen += nread
			if message == nil || readLen > dataLen {
				break
			}
			zlog.Infof("session receive msg, id: %d size: %d", message.GetMsgId(), message.GetMsgLen())
		}
	}
}

func (s *Session) startWriter() {
	zlog.Info("session writer start: ", s.sessionId)
}

func (s *Session) finalizer() {
	zlog.Info("session finalizer: ", s.sessionId)
}
