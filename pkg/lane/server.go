package lane

import (
	"maps"
	"slices"
	"sync"
	"sync/atomic"
	"time"

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

type (
	Server struct {
		accept        Acceptor
		sessions      sync.Map
		serdeBuilders map[string]serdeBuilder
		wc            loom.WaitClose

		heartbeatBuffer   []byte       // 心跳包只有一个类型, 没有具体的数据字段, 因此与serde无关
		idGenerator       atomic.Int64 // Session ID生成器，替代全局变量
		heartbeatInterval time.Duration
		kickInterval      time.Duration

		handlerLock   sync.RWMutex
		routeHandlers map[string]*HandlerItem
	}
)

func NewServer(accept Acceptor, opts ...ServerOption) *Server {
	// 默认值
	var options = serverOptions{
		SerdeBuilders:     map[string]serdeBuilder{},
		HeartbeatInterval: 3 * time.Second,
		KickInterval:      time.Minute,
	}

	// 初始化
	for _, opt := range opts {
		opt(&options)
	}

	var server = &Server{
		accept:            accept,
		serdeBuilders:     options.SerdeBuilders,
		heartbeatBuffer:   createCommonPackBuffer(serde.Packet{Route: convert.Bytes(serde.Heartbeat)}),
		heartbeatInterval: options.HeartbeatInterval,
		kickInterval:      options.KickInterval,
		routeHandlers:     make(map[string]*HandlerItem),
	}

	// 除默认支持JsonSerde外, 可额外添加ProtoSerde等支持
	maps.Copy(server.serdeBuilders, options.SerdeBuilders)

	loom.Go(server.goLoop)
	return server
}

func (my *Server) goLoop(later loom.Later) {
	var closeChan = my.wc.C()

	for {
		select {
		case conn := <-my.accept.GetLinkChan():
			// 外网环境是非常恶劣的, 有大量扫描器
			// 先写个心跳包检测一下链接的可用性, 如果失败了就无需建立session了
			if _, err := conn.Write(my.heartbeatBuffer); err != nil {
				logo.Info("heartbeat write failed, addr=%s, err=%v", conn.RemoteAddr(), err)
				_ = conn.Close() // 显式忽略：commonLink.Close()永不为error，仅设置isClosed标志
				continue
			}

			my.onNewSession(conn)
		case <-closeChan:
			return
		}
	}
}

func (my *Server) onNewSession(conn intern.Link) {
	var id = my.idGenerator.Add(1)
	var session = newServerSession(my, conn, id)
	var err = session.Handshake()
	if err != nil {
		logo.Info("session handshake failed, addr=%s, err=%v", conn.RemoteAddr(), err)
		return
	}

	my.sessions.Store(id, session)

	session.OnClosed(func() {
		my.sessions.Delete(id)
	})
}

// Listen 初始化完成后才启动Listen, 才可以接受session连接进来
func (my *Server) Listen() {
	my.accept.Listen()
}

func (my *Server) On(route string, handler HandlerFn) error {
	if route == "" {
		return TraceError("NilRoute")
	}

	if handler == nil {
		return TraceError("NilHandler", "route", route)
	}

	my.handlerLock.Lock()
	defer my.handlerLock.Unlock()

	if _, exists := my.routeHandlers[route]; exists {
		return TraceError("RouteAlreadyExists", "route", route)
	}

	my.routeHandlers[route] = &HandlerItem{
		Route:   route,
		Handler: handler,
	}

	return nil
}

func (my *Server) getHandlerByRoute(route string) *HandlerItem {
	my.handlerLock.RLock()
	defer my.handlerLock.RUnlock()

	var handler = my.routeHandlers[route]
	return handler
}

func createCommonPackBuffer(pack serde.Packet) []byte {
	var stream = &iox.OctetsStream{}
	var writer = iox.NewOctetsWriter(stream)
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var result = slices.Clone(buffer)

	return result
}
