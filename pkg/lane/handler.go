package lane

import "context"

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type HandlerFn func(ctx context.Context, input any) error

type HandlerItem struct {
	Route   string
	Handler HandlerFn
}
