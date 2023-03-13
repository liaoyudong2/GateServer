package iface

type ISession interface {
	GetSessionId() uint32     // 获取会话id
	Close()                   // 关闭
	SendMessage(msg IMessage) // 发送信息
	RawBuffer(buf []byte)     // 发送原始数据
}
