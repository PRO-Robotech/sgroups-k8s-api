package client

import (
	sgroupsv1 "github.com/PRO-Robotech/sgroups-proto/pkg/api/sgroups/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	Conn          *grpc.ClientConn
	Namespaces    sgroupsv1.SGroupsNamespaceAPIClient
	AddressGroups sgroupsv1.SGroupsAddressGroupsAPIClient
}

func Dial(addr string, opts ...grpc.DialOption) (*Client, error) {
	if len(opts) == 0 {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return newClient(conn), nil
}

func (c *Client) Close() error {
	if c == nil || c.Conn == nil {
		return nil
	}

	return c.Conn.Close()
}
