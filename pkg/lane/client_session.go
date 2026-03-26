package lane

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/got/iox"
	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ClientSession struct {
	id        int64
	writeLock sync.Mutex
	writer    *iox.OctetsWriter
	wc        loom.WaitClose
	serde     serde.Serde
	nonce     int32
	conn      net.Conn

	heartbeatInterval  time.Duration
	onHandShaken       func(bean serde.JsonHandshake)
	requestIdGenerator int32

	requestHandlers map[int32]ClientHandlerFn
	routeHandlers   map[string]ClientHandlerFn
}

func NewClientSession() *ClientSession {
	var my = &ClientSession{
		writer:            iox.NewOctetsWriter(&iox.OctetsStream{}),
		heartbeatInterval: time.Minute, // 初始给一个大一些的值, 防止client自己超时, 回头server会重置该值
		requestHandlers:   make(map[int32]ClientHandlerFn),
		routeHandlers:     make(map[string]ClientHandlerFn),
	}

	return my
}

func (my *ClientSession) Close() error {
	return my.wc.Close(func() error {
		var err = my.conn.Close()
		return err
	})
}

func (my *ClientSession) Connect(address string, opts ...ClientOption) error {
	var options = &clientOptions{
		serde:        &serde.JsonSerde{},
		tlsConfig:    nil,
		onHandShaken: nil,
	}

	for _, opt := range opts {
		opt(options)
	}

	my.serde = options.serde

	var conn net.Conn
	var err error
	if options.tlsConfig != nil {
		conn, err = tls.Dial("tcp", address, options.tlsConfig)
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return err
	}

	my.conn = conn
	my.onHandShaken = options.onHandShaken
	go my.goLoop()

	return nil
}

func (my *ClientSession) goHeartbeat() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var pack = serde.Packet{
		Route: convert.Bytes(serde.Heartbeat),
	}

	for !my.wc.IsClosed() {
		_ = my.sendPacket(pack)
		time.Sleep(my.heartbeatInterval)
	}
}

func (my *ClientSession) goLoop() {
	defer loom.DumpIfPanic()
	defer my.Close()

	var buffer = make([]byte, 1024)
	var stream = &iox.OctetsStream{}
	var reader = iox.NewOctetsReader(stream)

	for !my.wc.IsClosed() {
		_ = my.conn.SetReadDeadline(time.Now().Add(my.heartbeatInterval * 3))
		var num, err1 = my.conn.Read(buffer)
		if err1 != nil {
			my.onReadHandler(nil, err1)
			return
		}

		_ = stream.Write(buffer[:num])
		my.onReadHandler(reader, nil)
		stream.Tidy()
	}
}

func (my *ClientSession) onReadHandler(reader *iox.OctetsReader, err error) {
	if err != nil {
		logo.Info("close session(%d) by err=%q", my.id, err)
		_ = my.Close()
		return
	}

	if err1 := my.onReceivedData(reader); err1 != nil {
		logo.Info("close session(%d) by onReceivedData(), err=%q", my.id, err1)
		_ = my.Close()
		return
	}
}

func (my *ClientSession) onReceivedData(reader *iox.OctetsReader) error {
	var packets = serde.DecodePacket(reader)
	for _, pack := range packets {
		var err1 = my.onReceivedPacket(pack)
		if err1 != nil {
			return err1
		}
	}

	return nil
}

func (my *ClientSession) onReceivedPacket(pack serde.Packet) error {
	var route = convert.String(pack.Route)
	switch route {
	case serde.Handshake:
		return my.onReceivedHandshake(pack)
	case serde.Heartbeat:
		break
	case serde.Kick:
		return my.Close()
	default:
		return my.onReceivedUserdata(pack)
	}
	return nil
}

func (my *ClientSession) onReceivedHandshake(pack serde.Packet) error {
	var handshake serde.JsonHandshake
	var err = convert.FromJsonE(pack.Data, &handshake)
	if err != nil {
		return err
	}

	logo.JsonI("handshake", handshake)
	my.nonce = handshake.Nonce
	my.heartbeatInterval = time.Duration(handshake.Heartbeat) * time.Second
	my.id = handshake.SessionId
	my.handshakeRe()

	if my.onHandShaken != nil {
		my.onHandShaken(handshake)
	}

	// 启动heartbeat
	go my.goHeartbeat()
	return nil
}

func (my *ClientSession) handshakeRe() {
	var reply = serde.JsonHandshakeRe{
		Serde: "json",
	}

	var replyData = convert.ToJson(reply)
	var pack = serde.Packet{
		Route: convert.Bytes(serde.HandshakeRe),
		Data:  replyData,
	}

	_ = my.sendPacket(pack)
}

func (my *ClientSession) onReceivedUserdata(pack serde.Packet) error {
	var handler = my.fetchHandler(pack)
	if handler == nil {
		// 有些协议, 真不想处理, 就不设置handlers了. 通常只要有requestId, 就是故意不处理的
		if pack.RequestId == 0 {
			logo.Warn("no handler, route=%s, requestId=0, data=%s", convert.String(pack.Route), convert.String(pack.Data))
		}

		return nil
	}

	var hasError = len(pack.Code) > 0
	if hasError {
		var code = convert.String(pack.Code)
		var message = convert.String(pack.Data)
		var err = NewError(code, "%s", message)

		handler(my, nil, err)
	} else {
		handler(my, pack.Data, nil)
	}

	return nil
}

func (my *ClientSession) fetchHandler(pack serde.Packet) ClientHandlerFn {
	var requestId = pack.RequestId
	if requestId != 0 {
		if handler, ok := my.requestHandlers[requestId]; ok {
			delete(my.requestHandlers, requestId)
			return handler
		}
	} else {
		var route = convert.String(pack.Route)
		if handler, ok := my.routeHandlers[route]; ok {
			return handler
		}
	}

	return nil
}

func (my *ClientSession) Send(route string, v any) error {
	if route == "" {
		return ErrInvalidRoute
	}

	if my.serde == nil {
		return ErrNilSerde
	}

	if my.wc.IsClosed() || v == nil {
		return nil
	}

	var payload, err1 = serializeOrRaw(my.serde, v)
	if err1 != nil {
		return err1
	}

	var pack = serde.Packet{Route: convert.Bytes(route), Data: payload}
	var err2 = my.sendPacket(pack)
	return err2
}

func (my *ClientSession) Request(route string, request any, handler func(session *ClientSession, responseBytes []byte, err *Error)) error {
	if my.serde == nil {
		return ErrNilSerde
	}

	if route == "" || request == nil {
		return ErrInvalidArgument
	}

	var data, err = my.serde.Serialize(request)
	if err != nil {
		return err
	}

	my.requestIdGenerator++
	var requestId = my.requestIdGenerator
	var pack = serde.Packet{
		Route:     convert.Bytes(route),
		RequestId: requestId,
		Data:      data,
	}

	if handler != nil {
		my.requestHandlers[requestId] = handler
	}

	return my.sendPacket(pack)
}

func (my *ClientSession) On(route string, handler ClientHandlerFn) error {
	if route == "" {
		return TraceError("NilRoute")
	}

	if handler == nil {
		return TraceError("NilHandler", "route", route)
	}

	if _, exists := my.routeHandlers[route]; exists {
		return TraceError("RouteAlreadyExists", "route", route)
	}

	my.routeHandlers[route] = handler
	return nil
}

func (my *ClientSession) sendPacket(pack serde.Packet) error {
	my.writeLock.Lock()
	defer my.writeLock.Unlock()

	var writer = my.writer
	var stream = writer.Stream()
	stream.Reset()
	serde.EncodePacket(writer, pack)

	var buffer = stream.Bytes()
	var _, err = my.conn.Write(buffer)
	return err
}

func (my *ClientSession) Nonce() int32 {
	return my.nonce
}

func (my *ClientSession) Serde() serde.Serde {
	return my.serde
}
