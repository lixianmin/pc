package serde

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

const (
	Handshake   = "!1" // 连接建立后, server主动发送handshake, 必须使用json序列化
	HandshakeRe = "!2" // 链接建立后, client收到handshake后的回复协议, 指定client使用的序列化方法, 必须使用json序列化
	Heartbeat   = "!3" // client定期发送心跳
	Kick        = "!4" // server踢人
)
