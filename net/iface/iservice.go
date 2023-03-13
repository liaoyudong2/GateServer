package iface

// IService
// @Description: 服务接口
type IService interface {
	StartService(port int, version string) // 启动服务
	StopService()                          // 停止服务
	SetMaxSession(num int)                 // 设置最大连接数量
	GetSessionMgr() ISessionMgr            // 获取连接管理对象
}
