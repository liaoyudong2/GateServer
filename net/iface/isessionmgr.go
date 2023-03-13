package iface

type ISessionMgr interface {
	GetSessionCount() int                 // 会话数量
	AddSession(session ISession)          // 添加会话
	RemoveSession(sessionId uint32)       // 移除会话
	GetSession(sessionId uint32) ISession // 获取会话
	CleanSession()                        // 移除所有会话
}
