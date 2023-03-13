package zlog

var Ins = NewZLog(BitDefault)

func Flags() int {
	return Ins.Flags()
}

func ResetFlags(flag int) {
	Ins.ResetFlags(flag)
}

func AddFlag(flag int) {
	Ins.AddFlag(flag)
}

func SetLogPath(path string) {
	Ins.SetLogPath(path)
}

func SetLogConsole() {
	Ins.SetConsole(true)
}

func CloseDebug() {
	Ins.CloseDebug()
}

func OpenDebug() {
	Ins.OpenDebug()
}

func Debugf(format string, v ...interface{}) {
	Ins.Debugf(format, v...)
}

func Debug(v ...interface{}) {
	Ins.Debug(v...)
}

func Infof(format string, v ...interface{}) {
	Ins.Infof(format, v...)
}

// Info -
func Info(v ...interface{}) {
	Ins.Info(v...)
}

func Warnf(format string, v ...interface{}) {
	Ins.Warnf(format, v...)
}

func Warn(v ...interface{}) {
	Ins.Warn(v...)
}

func Errorf(format string, v ...interface{}) {
	Ins.Errorf(format, v...)
}

func Error(v ...interface{}) {
	Ins.Error(v...)
}

func Fatalf(format string, v ...interface{}) {
	Ins.Fatalf(format, v...)
}

func Fatal(v ...interface{}) {
	Ins.Fatal(v...)
}

func Panicf(format string, v ...interface{}) {
	Ins.Panicf(format, v...)
}

func Panic(v ...interface{}) {
	Ins.Panic(v...)
}

func Stack(v ...interface{}) {
	Ins.Stack(v...)
}

func init() {
	//因为StdZinxLog对象 对所有输出方法做了一层包裹，所以在打印调用函数的时候，比正常的logger对象多一层调用
	//一般的zinxLogger对象 callDepth=2, StdZinxLog的calldDepth=3
	Ins.callDepth = 3
}
