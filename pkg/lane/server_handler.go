package lane

import (
	"context"

	"github.com/lixianmin/got/convert"
	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ServerHandlerFn func(ctx context.Context, pack serde.Packet) error

type ServerHanderFnT[T any, U any] func(ctx context.Context, request T) (response U, err error)

func ServerHandler[T any, U any](handler ServerHanderFnT[T, U]) ServerHandlerFn {
	return func(ctx context.Context, input serde.Packet) error {
		var se = ctx.Value(keySession)
		if se == nil {
			return TraceError("NilSession")
		}

		var session = se.(*ServerSession)
		var serdeFn = session.Serde()
		if serdeFn == nil {
			return ErrNilSerde
		}

		var typed T
		if len(input.Data) > 0 {
			if err := serdeFn.Deserialize(input.Data, &typed); err != nil {
				return err
			}
		}

		var response, err = handler(ctx, typed)

		// 这个方法有可能从非session线程中回调回来, 因此要求my.serder必须是thread-safe的
		var payload []byte
		if err == nil {
			payload, err = serializeOrRaw(serdeFn, response)
		}

		var output = serde.Packet{
			Route:     input.Route,
			RequestId: input.RequestId,
		}

		if err == nil {
			output.Data = payload
		} else if err2, ok := err.(*Error); ok {
			output.Code = convert.Bytes(err2.Code)
			output.Data = convert.Bytes(err2.Message)
		} else {
			output.Code = convert.Bytes("PlainError")
			output.Data = convert.Bytes(err.Error())
		}

		return session.sendPacket(output)
	}
}
