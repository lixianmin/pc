package lane

import (
	"math/rand"

	"github.com/lixianmin/pc/pkg/lane/serde"
)

/********************************************************************
created:    2020-09-01
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

// keyType is a type for context keys
type keyType struct{}

var keyNonce = keyType{}
var keySession = keyType{}

// func GetSessionFromCtx(ctx context.Context) Session {
// 	var fetus = ctx.Value(keySession)
// 	if fetus == nil {
// 		logo.Warn("ctx doesn't contain the session")
// 		return nil
// 	}

// 	return fetus.(*sessionImpl)
// }

func serializeOrRaw(serde serde.Serde, v any) ([]byte, error) {
	if data, ok := v.([]byte); ok {
		return data, nil
	}

	var data, err = serde.Serialize(v)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func fetchNonce() int32 {
	// nonce一定不为0
	for {
		var nonce = rand.Int31()
		if nonce != 0 {
			return nonce
		}
	}
}
