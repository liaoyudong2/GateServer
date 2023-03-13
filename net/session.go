package net

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/liaoyudong2/GateServer/net/iface"
	"github.com/liaoyudong2/GateServer/zlog"
	"sync"
)

type Session struct {
	sessionId  uint32              // 会话ID
	conn       *websocket.Conn     // 连接对象
	closed     bool                // 是否已关闭
	lock       sync.RWMutex        // 读写锁
	exitChan   chan bool           // 关闭信号
	exitStr    string              // 退出原因
	writeChan  chan iface.IMessage // 写通道
	rawChan    chan []byte         // 原始写通道
	netStream  iface.IStream       // 解析
	sessionMgr iface.ISessionMgr   // 管理方
}

func NewSession(sessionId uint32, conn *websocket.Conn, sessionMgr iface.ISessionMgr) iface.ISession {
	session := &Session{
		sessionId:  sessionId,
		conn:       conn,
		closed:     false,
		exitChan:   make(chan bool, 1),
		writeChan:  make(chan iface.IMessage, 1),
		rawChan:    make(chan []byte, 1),
		netStream:  NewStream(),
		sessionMgr: sessionMgr,
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
	s.exitChan <- true
	zlog.Infof("session close, session id is [%d], reason: [%s]", s.sessionId, s.exitStr)
	s.sessionMgr.RemoveSession(s.sessionId)
}

func (s *Session) SendMessage(msg iface.IMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()

	zlog.Infof("session write message, msgId:%d, msgLen:%d", msg.GetMsgId(), msg.GetMsgLen())
	s.writeChan <- msg
}

func (s *Session) RawBuffer(buf []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	zlog.Infof("session write raw buffer")
	s.rawChan <- buf
}

func (s *Session) startReader() {
	zlog.Info("session reader start: ", s.sessionId)
	defer s.Close()

	for {
		msgType, data, err := s.conn.ReadMessage()
		if err != nil {
			zlog.Error("session read error: ", err)
			s.exitStr = err.Error()
			break
		}
		if msgType != websocket.BinaryMessage {
			zlog.Error("session read msg type error")
			s.exitStr = fmt.Sprintf("websocket msg type error, [%d]", msgType)
			break
		}
		readLen := 0
		dataLen := len(data)
		sessionShutdown := false
		for {
			nread, message, err := s.netStream.Unmarshal(data)
			if err != nil {
				zlog.Error("session net stream unmarshal error: ", err)
				sessionShutdown = true
				break
			}
			readLen += nread
			if message == nil || readLen > dataLen {
				break
			}
			zlog.Infof("session receive msg, id: %d size: %d", message.GetMsgId(), message.GetMsgLen())
		}
		if sessionShutdown {
			s.exitStr = fmt.Sprintf("session message unmarshal failed")
			break
		}
	}
	zlog.Info("session reader stop: ", s.sessionId)
}

func (s *Session) startWriter() {
	zlog.Info("session writer start: ", s.sessionId)

	select {
	case <-s.exitChan:
		break
	case msg, ok := <-s.writeChan:
		if ok {
			if err := s.conn.WriteMessage(websocket.BinaryMessage, s.netStream.Marshal(msg)); err != nil {
				zlog.Error("session write err: ", err)
			}
			zlog.Infof("session write message ok, msgId:%d, msgSize:%d", msg.GetMsgId(), msg.GetMsgLen())
		}
	case buf, ok := <-s.rawChan:
		if ok {
			if err := s.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
				zlog.Error("session write err: ", err)
			}
			zlog.Infof("session write raw buffer ok")
		}
	}
	zlog.Info("session writer stop: ", s.sessionId)
}
