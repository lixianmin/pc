package lane

import (
	"crypto/tls"

	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2020-08-29
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type clientOptions struct {
	serde        serde.Serde
	tlsConfig    *tls.Config
	onHandShaken func(bean serde.JsonHandshake)
}

type ClientOption func(options *clientOptions)

func WithSerde(serde serde.Serde) ClientOption {
	return func(opt *clientOptions) {
		opt.serde = serde
	}
}

func WithTlsConfig(config *tls.Config) ClientOption {
	return func(opt *clientOptions) {
		opt.tlsConfig = config
	}
}

func WithOnHandShaken(handler func(bean serde.JsonHandshake)) ClientOption {
	return func(opt *clientOptions) {
		opt.onHandShaken = handler
	}
}
