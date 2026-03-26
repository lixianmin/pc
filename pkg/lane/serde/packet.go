package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Packet struct {
	Route     []byte
	RequestId int32  // 请求的rid, 用于client请求时定位response的handler
	Code      []byte // error code
	Data      []byte // 如果有error code, 则Data是error message; 否则Data是数据payload
}

// JsonHandshake handshake必须使用json做序列化与反序列化
type JsonHandshake struct {
	Nonce     int32    `json:"nonce"`
	Heartbeat float32  `json:"heartbeat"` // 心跳间隔. 单位: 秒
	Serdes    []string `json:"serdes"`    // 服务器支持的序列化方法
	SessionId int64    `json:"sid"`       // session id
}

type JsonHandshakeRe struct {
	Serde string `json:"serde"`
}

type JsonRouteKind struct {
	Kind  int32  `json:"kind"`
	Route string `json:"route"`
}
