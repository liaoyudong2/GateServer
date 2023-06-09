package iface

type IMessage interface {
	GetMsgId() uint16   // 获取消息ID
	GetMsgLen() int     // 获取消息长度
	GetMsgData() []byte // 获取消息内容
}

type IStream interface {
	Unmarshal(data []byte) (int, IMessage, error) // 解包消息
	Marshal(msg IMessage) []byte                  // 打包消息
	SetMaxSize(size uint32)                       // 最大消息长度
}
