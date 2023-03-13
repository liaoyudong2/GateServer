package iface

// IService
// @Description: 服务接口
type IService interface {
	StartService(port int)                               // 启动服务
	StopService()                                        // 停止服务
	SetMaxSession(num int)                               // 设置最大连接数量
	GetSessionMgr() ISessionMgr                          // 获取连接管理对象
	SendMessageToSession(sessionId uint32, msg IMessage) // 发送消息
	RawBufferToSession(sessionId uint32, buf []byte)     // 发送原始数据给客户端
}
