package autostart

// Manager 开机自启动管理接口，定义注册、取消和查询操作
type Manager interface {
	// Enable 注册开机自启动
	Enable() error
	// Disable 取消开机自启动
	Disable() error
	// IsEnabled 查询是否已注册开机自启动
	IsEnabled() bool
}

// New 根据当前操作系统创建对应的自启动管理器
// appName 为注册时使用的应用名称标识
func New(appName string) Manager {
	return newManager(appName)
}
