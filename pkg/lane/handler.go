package lane

import (
	"context"

	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type innerHandlerFn func(ctx context.Context, input any) error

type HandlerFn[T any] func(ctx context.Context, input T) error

type HandlerItem struct {
	Route   string
	Handler innerHandlerFn
}

func wrapTypedHandler[T any](handler HandlerFn[T]) innerHandlerFn {
	return func(ctx context.Context, input any) error {
		var pack, ok = input.(serde.Packet)
		if !ok {
			return TraceError("InvalidPacketType")
		}

		var session = ctx.Value(keySession)
		if session == nil {
			return TraceError("NilSession")
		}

		var session1 = session.(*ServerSession)
		var serde = session1.serde
		if serde == nil {
			return ErrNilSerde
		}

		var typed T
		if len(pack.Data) > 0 {
			if err := serde.Deserialize(pack.Data, &typed); err != nil {
				return err
			}
		}

		return handler(ctx, typed)
	}
}
