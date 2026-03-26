package lane

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/lane/intern"
	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerSession struct {
	server     *Server
	writer     *iox.OctetsWriter
	writeLock  sync.Mutex
	id         int64
	link       intern.Link
	ctxValue   reflect.Value
	attachment *AttachmentImpl
	wc         loom.WaitClose
	serde      serde.Serde

	handlerLock      sync.Mutex
	onClosedHandlers []func()
}

func newServerSession(server *Server, link intern.Link, id int64) *ServerSession {
	var my = &ServerSession{
		server:     server,
		writer:     iox.NewOctetsWriter(&iox.OctetsStream{}),
		id:         id,
		link:       link,
		attachment: &AttachmentImpl{},
	}

	// 线上有大量的非法请求, 感觉是攻击, 先用Debug输出吧, 否则会生成大量无效日志
	logo.Info("create session(%d), addr=%s", my.id, my.link.RemoteAddr())
	var ctx = context.WithValue(context.Background(), keySession, my)
	my.attachment.Set(KeyContext, ctx)

	my.ctxValue = reflect.ValueOf(ctx)
	my.startGoLoop()

	// 这个设计, 可以保证finalizer被调用到, 但极大延长了对象在内存中存活的时间, 导致内存上涨很快. 外网扫描器很多, 有可能导致内存OOM
	//// 参考: https://zhuanlan.zhihu.com/p/76504936
	//runtime.SetFinalizer(my, func(w *sessionWrapper) {
	//	_ = w.Close()
	//})

	return my
}

func (my *ServerSession) startGoLoop() {
	go my.link.GoLoop(my.server.kickInterval, func(reader *iox.OctetsReader, err error) {
		if err != nil {
			logo.Info("close session(%d) by err, addr=%s, err=%q", my.id, my.link.RemoteAddr(), err)
			_ = my.Close()
			return
		}

		if err1 := my.onReceivedData(reader); err1 != nil {
			logo.Info("close session(%d) by onReceivedData(), addr=%s, err=%q", my.id, my.link.RemoteAddr(), err1)
			_ = my.Close()
			return
		}
	})
}

// Close 可以被多次调用，但只触发一次OnClosed事件
func (my *ServerSession) Close() error {
	return my.wc.Close(func() error {
		var err = my.link.Close()
		my.attachment.dispose()
		my.onClosed()
		return err
	})
}

func (my *ServerSession) onClosed() {
	my.handlerLock.Lock()
	defer my.handlerLock.Unlock()

	{
		for _, handler := range my.onClosedHandlers {
			handler()
		}
		my.onClosedHandlers = nil
	}
}

func (my *ServerSession) onReceivedData(reader *iox.OctetsReader) error {
	var packets = serde.DecodePacket(reader)
	for _, pack := range packets {
		if err := my.onReceivedPacket(pack); err != nil {
			return err
		}
	}

	return nil
}

func (my *ServerSession) onReceivedPacket(pack serde.Packet) error {
	var route = convert.String(pack.Route)
	switch route {
	case serde.Handshake:
		return nil
	case serde.HandshakeRe:
		return my.onReceivedHandshakeRe(pack)
	case serde.Heartbeat:
		// 现在server只有一个goroutine用于阻塞式读取网络数据，因此server缺少定时发送heartbeat的能力，因此采用client主动heartbeat而server回复的方案
		if _, err2 := my.link.Write(my.server.heartbeatBuffer); err2 != nil {
			return err2
		}
		return nil
	case serde.Kick:
		return my.Close()
	default:
		return my.onReceivedUserdata(pack)
	}
}

func (my *ServerSession) onReceivedHandshakeRe(input serde.Packet) error {
	var info serde.JsonHandshakeRe
	var err = convert.FromJsonE(input.Data, &info)
	if err != nil {
		return err
	}

	var serde = my.server.createSerde(info.Serde, my)
	if serde == nil {
		return NewError("InvalidSerde", "info.Serde=%s", info.Serde)
	}

	my.serde = serde
	// my.onEventHandShaken()
	return nil
}

func (my *ServerSession) onReceivedUserdata(input serde.Packet) error {
	// client发来的消息, 必须有handlerItem, 因此一定有kind才是合理的. server推送的消息可以没有kind
	var route = convert.String(input.Route)
	var handlerItem = my.server.getHandlerByRoute(route)
	if handlerItem == nil {
		return ErrEmptyHandler
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	var ctx = context.WithValue(context.Background(), keySession, my)
	handlerItem.Handler(ctx, input)

	return nil
}

func (my *ServerSession) Handshake() error {
	if my.wc.IsClosed() {
		return nil
	}

	var nonce = fetchNonce()
	var info = serde.JsonHandshake{
		Nonce:     nonce,
		Heartbeat: float32(my.server.heartbeatInterval.Seconds()),
		SessionId: my.id, // server的很多日志都是基于sid的, client打印一下这个值, 用于跟server配对
	}

	// all supported serde names
	for name := range my.server.serdeBuilders {
		info.Serdes = append(info.Serdes, name)
	}

	// handshake这个协议一定使用json去发, 后续的协议则可以替换为其它serde方法
	var data, err1 = convert.ToJsonE(info)
	if err1 != nil {
		return err1
	}

	my.Attachment().Set(keyNonce, nonce)
	var pack = serde.Packet{Route: convert.Bytes(serde.Handshake), Data: data}
	var err2 = my.sendPacket(pack)

	if err2 != nil {
		_ = my.Close()
	}

	return err2
}

func (my *ServerSession) sendPacket(pack serde.Packet) error {
	my.writeLock.Lock()
	defer my.writeLock.Unlock()

	var writer = my.writer
	var stream = writer.Stream()
	stream.Reset()
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var _, err = my.link.Write(buffer)
	return err
}

// OnClosed 需要保证OnClosed事件在任何情况下都会有且仅有一次触发：无论是主动断开，还是意外断开链接；无论client端有没有因为网络问题收到回复消息
func (my *ServerSession) OnClosed(handler func()) {
	if handler != nil {
		my.handlerLock.Lock()
		defer my.handlerLock.Unlock()

		my.onClosedHandlers = append(my.onClosedHandlers, handler)
	}
}

func (my *ServerSession) Send(route string, v any) error {
	if route == "" {
		return ErrInvalidRoute
	}

	if my.wc.IsClosed() {
		return nil
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	var pack = serde.Packet{Route: convert.Bytes(route)}

	var err2, isError2 = v.(error)
	if !isError2 {
		var payload, err3 = serializeOrRaw(my.serde, v)
		if err3 != nil {
			return err3
		}

		pack.Data = payload
	} else if err4, ok := v.(*Error); ok {
		pack.Code = convert.Bytes(err4.Code)
		pack.Data = convert.Bytes(err4.Message)
	} else {
		pack.Code = convert.Bytes("PlainError")
		pack.Data = convert.Bytes(err2.Error())
	}

	var err5 = my.sendPacket(pack)
	return err5
}

// Id 全局唯一id
func (my *ServerSession) Id() int64 {
	return my.id
}

func (my *ServerSession) RemoteAddr() net.Addr {
	return my.link.RemoteAddr()
}

func (my *ServerSession) Attachment() Attachment {
	return my.attachment
}

func (my *ServerSession) Nonce() int32 {
	return my.attachment.Int32(keyNonce)
}

func (my *ServerSession) Serde() serde.Serde {
	return my.serde
}
