package lane

import (
	"context"
	"testing"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/lane/epoll"
	"github.com/lixianmin/pc/pkg/lane/serde"
)

func TestServer(t *testing.T) {
	const address = ":3456"
	var acceptor = epoll.NewTcpAcceptor(address)
	var server = NewServer(acceptor, WithSerdeBuilder("json", func(session *ServerSession) serde.Serde {
		return &serde.JsonSerde{}
	}))

	server.On("hello", func(ctx context.Context, input any) error {
		logo.JsonI("hello", "world")
		return nil
	})

	loom.Go(func(later loom.Later) {
		server.Listen()
	})

	var client = NewClientSession()
	client.Connect(address, WithSerde(&serde.JsonSerde{}), WithOnHandShaken(func(bean *serde.JsonHandshake) {
		logo.JsonI("handshaken", bean)
	}))

	select {}
}
