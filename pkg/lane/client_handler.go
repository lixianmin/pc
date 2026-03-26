package lane

/********************************************************************
created:    2026-03-26
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

type ClientHandlerFn func(session *ClientSession, responseBytes []byte, err *Error)

type ClientHandlerFnT[T any] func(response T, err *Error)

func ClientHandler[T any](handler ClientHandlerFnT[T]) ClientHandlerFn {
	return func(session *ClientSession, responseBytes []byte, err *Error) {
		var serde = session.Serde()
		if serde == nil || len(responseBytes) == 0 {
			return
		}

		var response T
		if err1 := serde.Deserialize(responseBytes, &response); err1 != nil {
			return
		}

		handler(response, err)
	}
}
