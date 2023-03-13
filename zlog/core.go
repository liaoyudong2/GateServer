package zlog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	LogMaxBuf = 1024 * 1024
)

// 日志头部信息标记位，采用bitmap方式，用户可以选择头部需要哪些标记位被打印
const (
	BitDate         = 1 << iota                            //日期标记位  2019/01/23
	BitTime                                                //时间标记位  01:23:12
	BitMicroSeconds                                        //微秒级标记位 01:23:12.111222
	BitLongFile                                            //完整文件名称 /home/go/src/server.go
	BitShortFile                                           //最后文件名   server.go
	BitLevel                                               //当前日志级别： 0(Debug), 1(Info), 2(Warn), 3(Error), 4(Panic), 5(Fatal)
	BitStdFlag      = BitDate | BitTime                    //标准头部日志格式
	BitDefault      = BitLevel | BitShortFile | BitStdFlag //默认日志头部格式
)

// 日志级别
const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
	LogPanic
	LogFatal
)

// 日志级别对应的显示字符串
var levels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
	"[PANIC]",
	"[FATAL]",
}

type ZLoggerWriter struct {
	path string    // 日志路径
	date string    // 日期字符串
	out  io.Writer // 日志输出的文件描述符
}

type ZLoggerCore struct {
	console    *os.File      // 控制台模式
	mu         sync.Mutex    //确保多协程读写文件，防止文件内容混乱，做到协程安全
	flag       int           //日志标记位
	writer     ZLoggerWriter // 日志输出
	buf        bytes.Buffer  //输出的缓冲区
	file       *os.File      //当前日志绑定的输出文件
	debugClose bool          //是否打印调试debug信息
	callDepth  int           //获取日志文件名和代码上述的runtime.Call 的函数调用层数
}

// NewZLog
//
//	@Description: 创建一个日志
//	@param flag 当前日志头部信息的标记位
//	@return *ZLoggerCore
func NewZLog(flag int) *ZLoggerCore {
	//默认 debug打开， calledDepth深度为2,ZLogger对象调用日志打印方法最多调用两层到达output函数
	zlog := &ZLoggerCore{
		writer: ZLoggerWriter{
			date: time.Now().Format(time.DateOnly),
			out:  os.Stdout,
		},
		flag:       flag,
		file:       nil,
		debugClose: false,
		callDepth:  2,
	}
	//设置log对象 回收资源 析构方法(不设置也可以，go的Gc会自动回收，强迫症没办法)
	runtime.SetFinalizer(zlog, CleanZLog)
	return zlog
}

// CleanZLog
//
//	@Description: 回收日志处理
//	@param log
func CleanZLog(log *ZLoggerCore) {
	log.closeFile()
}

// formatHeader
//
//	@Description: 制作当条日志数据的 格式头信息
//	@receiver log
//	@param t 时间
//	@param file 来源代码文件名
//	@param line 来源代码文件行数
//	@param level 日志等级
func (log *ZLoggerCore) formatHeader(t time.Time, file string, line int, level int) {
	var buf = &log.buf

	//已经设置了时间相关的标识位,那么需要加时间信息在日志头部
	if log.flag&(BitDate|BitTime|BitMicroSeconds) != 0 {
		//日期位被标记
		if log.flag&BitDate != 0 {
			year, month, day := t.Date()
			buf.WriteByte('[')
			itoa(buf, year, 4)
			buf.WriteByte('-') // "[2019-"
			itoa(buf, int(month), 2)
			buf.WriteByte('-') // "[2019-04-"
			itoa(buf, day, 2)
			buf.WriteByte(' ') // "[2019-04-11 "
		}

		//时钟位被标记
		if log.flag&(BitTime|BitMicroSeconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			buf.WriteByte(':') // "11:"
			itoa(buf, min, 2)
			buf.WriteByte(':') // "11:15:"
			itoa(buf, sec, 2)  // "11:15:33"
			//微秒被标记
			if log.flag&BitMicroSeconds != 0 {
				buf.WriteByte('.')
				itoa(buf, t.Nanosecond()/1e3, 6) // "11:15:33.123123
			}
			buf.WriteByte(']')
			buf.WriteByte(' ')
		}

		// 日志级别位被标记
		if log.flag&BitLevel != 0 {
			buf.WriteString(levels[level])
		}
		buf.WriteByte(' ')

		//日志当前代码调用文件名名称位被标记
		if log.flag&(BitShortFile|BitLongFile) != 0 {
			//短文件名称
			if log.flag&BitShortFile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						//找到最后一个'/'之后的文件名称
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			buf.WriteString(file)
			buf.WriteByte(':')
			itoa(buf, line, -1) //行数
			buf.WriteString(": ")
		}
	}
}

// OutPut
//
//	@Description: 输出日志文件,原方法
//	@receiver log
//	@param level 日志等级
//	@param s 日志源内容(未加工)
//	@return error
func (log *ZLoggerCore) OutPut(level int, s string) error {

	now := time.Now() // 得到当前时间
	var file string   //当前调用日志接口的文件名称
	var line int      //当前代码行数
	log.mu.Lock()
	defer log.mu.Unlock()

	if log.flag&(BitShortFile|BitLongFile) != 0 {
		log.mu.Unlock()
		var ok bool
		//得到当前调用者的文件名称和执行到的代码行数
		_, file, line, ok = runtime.Caller(log.callDepth)
		if !ok {
			file = "unknown-file"
			line = 0
		}
		log.mu.Lock()
	}

	//清零buf
	log.buf.Reset()
	//写日志头
	log.formatHeader(now, file, line, level)
	//写日志内容
	log.buf.WriteString(s)
	//补充回车
	if len(s) > 0 && s[len(s)-1] != '\n' {
		log.buf.WriteByte('\n')
	}

	//将填充好的buf 写到IO输出上
	return log.flushBuffer()
}

// flushBuffer
//
//	@Description: 写入文件
//	@receiver log
//	@return error
func (log *ZLoggerCore) flushBuffer() error {
	log.openFile()
	if log.writer.out != nil {
		if log.console != nil {
			_, _ = log.console.Write(log.buf.Bytes())
		}
		//将填充好的buf 写到IO输出上
		_, err := log.writer.out.Write(log.buf.Bytes())
		return err
	}
	return errors.New("writer out fd is nil")
}

// openFile
//
//	@Description: 打开日志文件输出(自动判定日期进行拆分)
//	@receiver log
func (log *ZLoggerCore) openFile() {
	dateStr := time.Now().Format(time.DateOnly)
	if log.file == nil || log.writer.date != dateStr {
		log.closeFile()

		var file *os.File

		log.writer.date = dateStr
		//创建日志文件夹
		_ = mkdirLog(log.writer.path)

		fullPath := log.writer.path + "/" + log.writer.date + ".log"
		if log.checkFileExist(fullPath) {
			//文件存在，打开
			file, _ = os.OpenFile(fullPath, os.O_APPEND|os.O_RDWR, 0644)
		} else {
			//文件不存在，创建
			file, _ = os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
		}
		log.file = file
		log.writer.out = file
	}
}

// closeFile
//
//	@Description: 关闭日志绑定的文件
//	@receiver log
func (log *ZLoggerCore) closeFile() {
	if log.file != nil {
		_ = log.file.Close()
		log.file = nil
		log.writer.out = os.Stderr
	}
}

// SetLogPath
//
//	@Description: 设置日志路径
//	@receiver log
//	@param path 路径
func (log *ZLoggerCore) SetLogPath(path string) {
	log.mu.Lock()
	defer log.mu.Unlock()

	log.writer.path = path
}

func (log *ZLoggerCore) SetConsole(stat bool) {
	if stat {
		log.console = os.Stdout
	} else {
		log.console = nil
	}
}

func (log *ZLoggerCore) Debugf(format string, v ...interface{}) {
	if log.debugClose == true {
		return
	}
	_ = log.OutPut(LogDebug, fmt.Sprintf(format, v...))
}

func (log *ZLoggerCore) Debug(v ...interface{}) {
	if log.debugClose == true {
		return
	}
	_ = log.OutPut(LogDebug, fmt.Sprintln(v...))
}

func (log *ZLoggerCore) Infof(format string, v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintf(format, v...))
}

func (log *ZLoggerCore) Info(v ...interface{}) {
	_ = log.OutPut(LogInfo, fmt.Sprintln(v...))
}

func (log *ZLoggerCore) Warnf(format string, v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintf(format, v...))
}

func (log *ZLoggerCore) Warn(v ...interface{}) {
	_ = log.OutPut(LogWarn, fmt.Sprintln(v...))
}

func (log *ZLoggerCore) Errorf(format string, v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintf(format, v...))
}

func (log *ZLoggerCore) Error(v ...interface{}) {
	_ = log.OutPut(LogError, fmt.Sprintln(v...))
}

func (log *ZLoggerCore) Fatalf(format string, v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (log *ZLoggerCore) Fatal(v ...interface{}) {
	_ = log.OutPut(LogFatal, fmt.Sprintln(v...))
	os.Exit(1)
}

func (log *ZLoggerCore) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *ZLoggerCore) Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	_ = log.OutPut(LogPanic, s)
	panic(s)
}

func (log *ZLoggerCore) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, LogMaxBuf)
	n := runtime.Stack(buf, true) //得到当前堆栈信息
	s += string(buf[:n])
	s += "\n"
	_ = log.OutPut(LogError, s)
}

func (log *ZLoggerCore) Flags() int {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.flag
}

func (log *ZLoggerCore) ResetFlags(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag = flag
}

func (log *ZLoggerCore) AddFlag(flag int) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.flag |= flag
}

func (log *ZLoggerCore) CloseDebug() {
	log.debugClose = true
}

func (log *ZLoggerCore) OpenDebug() {
	log.debugClose = false
}

func (log *ZLoggerCore) checkFileExist(filename string) bool {
	exist := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func mkdirLog(dir string) (e error) {
	_, er := os.Stat(dir)
	b := er == nil || os.IsExist(er)
	if !b {
		if err := os.MkdirAll(dir, 0775); err != nil {
			if os.IsPermission(err) {
				e = err
			}
		}
	}
	return
}

// 将一个整形转换成一个固定长度的字符串，字符串宽度应该是大于0的
// 要确保buffer是有容量空间的
func itoa(buf *bytes.Buffer, i int, wID int) {
	var u = uint(i)
	if u == 0 && wID <= 1 {
		buf.WriteByte('0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wID > 0; u /= 10 {
		bp--
		wID--
		b[bp] = byte(u%10) + '0'
	}

	// avoID slicing b to avoID an allocation.
	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}
