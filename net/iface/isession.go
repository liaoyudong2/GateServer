package iface

type ISession interface {
	GetSessionId() uint32 // 获取会话id
	Close()               // 关闭
}
