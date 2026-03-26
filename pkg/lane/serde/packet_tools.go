package serde

import (
	"errors"
	"io"

	"github.com/lixianmin/got/iox"
)

/********************************************************************
created:    2023-06-05
author:     lixianmin

Copyright (C) - All Rights Reserved
*********************************************************************/

func EncodePacket(writer *iox.OctetsWriter, pack Packet) {
	_ = writer.WriteBytes(pack.Route)
	_ = writer.Write7BitEncodedInt(pack.RequestId)
	_ = writer.WriteBytes(pack.Code)
	_ = writer.WriteBytes(pack.Data)
}

func DecodePacket(reader *iox.OctetsReader) []Packet {
	var packets []Packet = nil
	var stream = reader.Stream()

	for {
		var lastPosition = stream.Position()

		var route, err = reader.ReadBytes()
		if errors.Is(err, iox.ErrNotEnoughData) {
			rewindStream(stream, lastPosition)
			return packets
		}

		requestId, err := reader.Read7BitEncodedInt()
		if errors.Is(err, iox.ErrNotEnoughData) {
			rewindStream(stream, lastPosition)
			return packets
		}

		code, err := reader.ReadBytes()
		if errors.Is(err, iox.ErrNotEnoughData) {
			rewindStream(stream, lastPosition)
			return packets
		}

		data, err := reader.ReadBytes()
		if errors.Is(err, iox.ErrNotEnoughData) {
			rewindStream(stream, lastPosition)
			return packets
		}

		var pack = Packet{Route: route, RequestId: requestId, Code: code, Data: data}
		packets = append(packets, pack)
	}
}

func rewindStream(stream *iox.OctetsStream, position int) {
	_, _ = stream.Seek(int64(position), io.SeekStart)
}
