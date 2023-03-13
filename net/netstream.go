package net

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/liaoyudong2/GateServer/net/iface"
)

// HeaderSize 头部固定大小
const HeaderSize = 10

type Message struct {
	msgSize uint32 // 消息长度
	msgId   uint16 // 消息id
	reserve uint32 // 保留字段(SessionID)
	msgData []byte // 消息内容
}

func (m Message) GetMsgId() uint16 {
	return m.msgId
}

func (m Message) GetMsgLen() int {
	if m.msgData != nil {
		return len(m.msgData)
	}
	return 0
}

func (m Message) GetMsgData() []byte {
	return m.msgData
}

type Stream struct {
	message Message // 头部信息
	buffer  []byte  // 不足信息时的头部缓冲区
	maxSize uint32  // 最大消息长度限制
}

func NewStream() iface.IStream {
	stream := &Stream{
		message: Message{
			msgSize: 0,
			msgId:   0,
			reserve: 0,
			msgData: make([]byte, 0x1000), // 默认分配容量
		},
		buffer:  make([]byte, HeaderSize),
		maxSize: 0x10000,
	}
	stream.cleanMessage()
	return stream
}

// cleanMessage
//
//	@Description: 清空消息记录
//	@receiver s
func (s *Stream) cleanMessage() {
	s.message.msgSize = 0
	s.message.msgId = 0
	s.message.reserve = 0
	s.message.msgData = s.message.msgData[:0]
	s.buffer = s.buffer[:0]
}

// parseHeader
//
//	@Description: 解析头部信息
//	@receiver s
//	@param data 流数据
//	@param mode 解析模式
func (s *Stream) parseHeader(data []byte) {
	buf := bytes.NewReader(data)
	_ = binary.Read(buf, binary.BigEndian, &s.message.msgSize)
	_ = binary.Read(buf, binary.BigEndian, &s.message.msgId)
	_ = binary.Read(buf, binary.BigEndian, &s.message.reserve)
}

func (s *Stream) Unmarshal(data []byte) (int, iface.IMessage, error) {
	readLen := 0
	dataLen := len(data)
	bufLen := len(s.buffer)
	headerSize := HeaderSize
	if bufLen > 0 {
		// 缺多少头部
		lessLen := headerSize - bufLen
		s.buffer = append(s.buffer, data[:lessLen]...)
		readLen += lessLen
		if dataLen < lessLen {
			// 头部长度不足, 下一次继续
			return readLen, nil, nil
		}
		// 解析头部
		s.parseHeader(s.buffer)
	} else {
		// 确认长度
		if dataLen < headerSize {
			s.buffer = append(s.buffer, data...)
			// 头部长度不足, 下一次继续
			return dataLen, nil, nil
		}
		s.parseHeader(data[:headerSize])
		readLen += headerSize
	}
	if s.message.msgSize > s.maxSize {
		s.cleanMessage()
		return 0, nil, errors.New(fmt.Sprintf("message size overflow, limit size is %d", s.maxSize))
	}
	dataLen -= readLen
	// 解析body
	msgLen := len(s.message.msgData)
	if msgLen > 0 {
		// 总容量是否足够
		if uint32(msgLen+dataLen) >= s.message.msgSize {
			// 拼接
			message := &Message{
				msgSize: s.message.msgSize,
				msgId:   s.message.msgId,
				reserve: s.message.reserve,
				msgData: make([]byte, s.message.msgSize),
			}
			lessLen := int(s.message.msgSize) - msgLen
			message.msgData = append(s.message.msgData, data[readLen:readLen+lessLen]...)
			readLen += lessLen
			// 清空
			s.cleanMessage()
			return readLen, message, nil
		}
		// 合并
		s.message.msgData = append(s.message.msgData, data[readLen:]...)
		// body信息不足
		return dataLen, nil, nil
	} else {
		if uint32(dataLen) >= s.message.msgSize {
			message := &Message{
				msgSize: s.message.msgSize,
				msgId:   s.message.msgId,
				reserve: s.message.reserve,
				msgData: make([]byte, s.message.msgSize),
			}
			message.msgData = bytes.Clone(data[readLen : readLen+int(s.message.msgSize)])
			readLen += int(s.message.msgSize)
			return readLen, message, nil
		}
		// 合并
		s.message.msgData = append(s.message.msgData, data[readLen:]...)
		readLen += dataLen
		return readLen, nil, nil
	}
}

func (s *Stream) Marshal(msg iface.IMessage) []byte {
	msgLen := HeaderSize + msg.GetMsgLen()
	buffer := bytes.NewBuffer(make([]byte, msgLen))
	_ = binary.Write(buffer, binary.BigEndian, msg.GetMsgLen())
	_ = binary.Write(buffer, binary.BigEndian, uint32(msg.GetMsgId()))
	_ = binary.Write(buffer, binary.BigEndian, uint32(0))
	if msg.GetMsgLen() > 0 {
		_ = binary.Write(buffer, binary.BigEndian, msg.GetMsgData())
	}
	return buffer.Bytes()
}

func (s *Stream) SetMaxSize(size uint32) {
	s.maxSize = size
}
