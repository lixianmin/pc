package lane

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/lane/intern"
	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2020-08-28
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type sessionWrapper struct {
	*sessionImpl
}

type sessionImpl struct {
	manager    *Manager
	writer     *iox.OctetsWriter
	writeLock  sync.Mutex
	id         int64
	link       intern.Link
	ctxValue   reflect.Value
	attachment *AttachmentImpl
	wc         loom.WaitClose
	serde      serde.Serde
	routeKinds map[string]int32

	handlerLock          sync.Mutex
	onHandShakenHandlers []func()
	onClosedHandlers     []func()
	echoChan             chan func()
	inReceiveLoop        int32 // 原子标记，用于检测是否在receive线程中调用Echo()
}

func newSession(manager *Manager, link intern.Link, id int64) Session {
	var routeKinds = manager.CloneRouteKinds()
	var my = &sessionWrapper{&sessionImpl{
		manager:    manager,
		writer:     iox.NewOctetsWriter(&iox.OctetsStream{}),
		id:         id,
		link:       link,
		attachment: &AttachmentImpl{},
		routeKinds: routeKinds,
		echoChan:   make(chan func(), 4),
	}}

	// 线上有大量的非法请求, 感觉是攻击, 先用Debug输出吧, 否则会生成大量无效日志
	logo.Info("create session(%d), addr=%s", my.id, my.link.RemoteAddr())
	var ctx = context.WithValue(context.Background(), keySession, my.sessionImpl)
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

// Close 可以被多次调用，但只触发一次OnClosed事件
func (my *sessionImpl) Close() error {
	return my.wc.Close(func() error {
		var err = my.link.Close()
		my.attachment.dispose()
		my.onEventClosed()
		return err
	})
}

func (my *sessionImpl) onEventClosed() {
	my.handlerLock.Lock()
	defer my.handlerLock.Unlock()
	{
		for _, handler := range my.onClosedHandlers {
			handler()
		}
		my.onClosedHandlers = nil
	}
}

func (my *sessionImpl) OnHandShaken(handler func()) {
	if handler != nil {
		my.handlerLock.Lock()
		defer my.handlerLock.Unlock()

		my.onHandShakenHandlers = append(my.onHandShakenHandlers, handler)
	}
}

// OnClosed 需要保证OnClosed事件在任何情况下都会有且仅有一次触发：无论是主动断开，还是意外断开链接；无论client端有没有因为网络问题收到回复消息
func (my *sessionImpl) OnClosed(handler func()) {
	if handler != nil {
		my.handlerLock.Lock()
		defer my.handlerLock.Unlock()

		my.onClosedHandlers = append(my.onClosedHandlers, handler)
	}
}

// Id 全局唯一id
func (my *sessionImpl) Id() int64 {
	return my.id
}

func (my *sessionImpl) RemoteAddr() net.Addr {
	return my.link.RemoteAddr()
}

func (my *sessionImpl) Attachment() Attachment {
	return my.attachment
}

func (my *sessionImpl) Nonce() int32 {
	return my.attachment.Int32(keyNonce)
}
