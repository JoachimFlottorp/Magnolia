package grpc

import (
	"context"

	"github.com/JoachimFlottorp/yeahapi/protobuf/collector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Instance struct {
	conn 			*grpc.ClientConn
	chattersClient 	collector.ChattersClient
}

// TODO keep connection alive

func NewInstance(addr string) (*Instance, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := collector.NewChattersClient(conn)

	i := &Instance{
		conn: conn,
		chattersClient: c,
	}

	return i, nil
}

func (i *Instance) ChatterGet(ctx context.Context, channel string) (*collector.ChatterResponse, error) {
	return i.chattersClient.GetChatters(ctx, &collector.ChatterRequest{
		Login: channel,
	})
}
