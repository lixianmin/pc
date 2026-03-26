package lane

import "github.com/lixianmin/pc/pkg/lane/serde"

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type Session interface {
	Nonce() int32
	Serde() serde.Serde
	Send(route string, v any) error
}
