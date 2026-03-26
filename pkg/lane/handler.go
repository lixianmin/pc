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

type HandlerFn func(ctx context.Context, input any) error

type HandlerFnTyped[T any] func(ctx context.Context, input T) error

type HandlerItem struct {
	Route   string
	Handler HandlerFn
}

func TypedHandler[T any](handler HandlerFnTyped[T]) HandlerFn {
	return func(ctx context.Context, input any) error {
		var pack, ok = input.(serde.Packet)
		if !ok {
			return TraceError("InvalidPacketType")
		}

		var session = ctx.Value(keySession)
		if session == nil {
			return TraceError("NilSession")
		}

		var session1 = session.(Session)
		var serde = session1.Serde()
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
