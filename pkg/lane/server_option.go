package lane

import (
	"time"

	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2020-09-30
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type serdeBuilder func(session *ServerSession) serde.Serde

type serverOptions struct {
	SerdeBuilders     map[string]serdeBuilder // 支持的serde列表
	HeartbeatInterval time.Duration           // heartbeat间隔
	KickInterval      time.Duration           // 因为玩家可能切游戏到后台很久去做其它的事情, 因此这个值必须要大一些, 太短很容易被服务器踢的, 参考link_common中的resetReadDeadline
}

type ServerOption func(*serverOptions)

func WithSerdeBuilder(name string, builder serdeBuilder) ServerOption {
	return func(options *serverOptions) {
		if name != "" && builder != nil {
			options.SerdeBuilders[name] = builder
		}
	}
}

func WithHeartbeatInterval(interval time.Duration) ServerOption {
	return func(options *serverOptions) {
		if interval > 0 {
			options.HeartbeatInterval = interval
		}
	}
}

func WithKickInterval(interval time.Duration) ServerOption {
	return func(options *serverOptions) {
		if interval > 0 {
			options.KickInterval = interval
		}
	}
}
