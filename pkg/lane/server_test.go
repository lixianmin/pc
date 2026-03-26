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

	type HelloRequest struct {
		Name string
	}

	type HelloResponse struct {
		Message string
	}

	server.On("hello", ServerHandler(func(ctx context.Context, request *HelloRequest) (*HelloResponse, error) {
		logo.JsonI("server_hello_response", request.Name)
		// var session = GetSessionFromCtx(ctx)
		// session.Send("hello", HelloResponse{Message: "hi, " + request.Name})
		return &HelloResponse{Message: "hi, " + request.Name}, nil
	}))

	loom.Go(func(later loom.Later) {
		server.Listen()
	})

	var client = NewClientSession()
	client.Connect(address, WithSerde(&serde.JsonSerde{}), WithOnHandShaken(func(handshake serde.JsonHandshake) {
		logo.JsonI("handshake", handshake)

		client.On("hello", ClientHandler(func(response *HelloResponse, err *Error) {
			if err != nil {
				logo.Warn("client handler error:", err)
				return
			}

			logo.JsonI("client_hello_response", response.Message)
		}))

		var request = HelloRequest{Name: "panda"}
		client.Send("hello", request)

		request.Name = "tiger"
		client.Send("hello", &request)

		request.Name = "kitten"
		client.Request("hello", request, ClientHandler(func(response *HelloResponse, err *Error) {
			if err != nil {
				logo.Error("request error:", err)
				return
			}

			logo.JsonI("response", response)
		}))
	}))

	select {}
}
