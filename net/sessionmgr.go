package net

import (
	"github.com/liaoyudong2/GateServer/net/iface"
	"github.com/liaoyudong2/GateServer/zlog"
	"sync"
)

type SessionMgr struct {
	sessions map[uint32]iface.ISession // 会话管理
	lock     sync.RWMutex              // 加锁
}

func NewSessionMgr(maxSession int) iface.ISessionMgr {
	return &SessionMgr{
		sessions: make(map[uint32]iface.ISession, maxSession),
	}
}

func (s *SessionMgr) GetSessionCount() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.sessions)
}

func (s *SessionMgr) AddSession(session iface.ISession) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if session, ok := s.sessions[session.GetSessionId()]; ok {
		zlog.Errorf("add session error: repeated session id: %v", session.GetSessionId())
	}
	s.sessions[session.GetSessionId()] = session
	zlog.Infof("session manager: [ADD] count:%d", len(s.sessions))
}

func (s *SessionMgr) RemoveSession(sessionId uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.sessions, sessionId)
	zlog.Infof("session manager: [REMOVE] count:%d", len(s.sessions))
}

func (s *SessionMgr) GetSession(sessionId uint32) iface.ISession {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if session, ok := s.sessions[sessionId]; ok {
		return session
	}
	return nil
}

func (s *SessionMgr) CleanSession() {
	for _, session := range s.sessions {
		session.Close()
	}
}
