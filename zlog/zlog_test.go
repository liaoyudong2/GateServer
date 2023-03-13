package zlog_test

import (
	"github.com/liaoyudong2/GateServer/zlog"
	"testing"
)

func TestStdZLog(t *testing.T) {

	//测试 默认debug输出
	zlog.Debug("zinx debug content1")
	zlog.Debug("zinx debug content2")

	zlog.Debugf(" zinx debug a = %d\n", 10)

	//设置log标记位，加上长文件名称 和 微秒 标记
	zlog.ResetFlags(zlog.BitDate | zlog.BitLongFile | zlog.BitLevel)
	zlog.Info("zinx info content")

	//添加标记位
	zlog.AddFlag(zlog.BitShortFile | zlog.BitTime)
	zlog.Stack(" Zinx Stack! ")

	//设置日志写入文件
	zlog.SetLogPath("./log/GateSrv")
	zlog.Debug("===> zinx debug content ~~666")
	zlog.Debug("===> zinx debug content ~~888")
	zlog.Error("===> zinx Error!!!! ~~~555~~~")

	//关闭debug调试
	zlog.CloseDebug()
	zlog.Debug("===> 我不应该出现~！")
	zlog.Debug("===> 我不应该出现~！")
	zlog.Error("===> zinx Error  after debug close !!!!")
}

func TestZLogger(t *testing.T) {
}
