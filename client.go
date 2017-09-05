package dgogm

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/dgraph/client"
	"google.golang.org/grpc"
)

// This struct holds the reference for dgraph client connection
// Dgorm supports using existing dgraph.Client or it can initiate the connection on it's own
type Dgraph struct {
	Addresses []string
	conns     []*grpc.ClientConn
	client    *client.Dgraph
}

// This function connects to the underlying grpc server and creates dgraph connections
// Dgraph 0.8.1 creates a client level cache, this function will create that file in
// os.TempDir()/dgraph/<timestamp_of_connection>
// If you require that file to be reused use ConnectWithClientDir
func Connect(addresses []string) (*Dgraph, error) {
	var err error
	d := new(Dgraph)
	d.Addresses = addresses
	for _, address := range addresses {
		c := context.Background()
		c, _ = context.WithTimeout(c, time.Second*30)
		conn, err := grpc.DialContext(c, address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		d.conns = append(d.conns, conn)
	}
	// This generates a unique folder structure for maintaining client cache
	d.client = client.NewDgraphClient(d.conns, client.DefaultOptions, fmt.Sprintf("%s/dgraph/%s", os.TempDir(), time.Now().UTC().String()))
	return d, err
}

// This function connects to the underlying grpc server and creates dgraph connections
// It will use the provided clientDir as the client dir in the connection
func ConnectWithClientDir(addresses []string, clientDir string) (*Dgraph, error) {
	var err error
	d := new(Dgraph)
	d.Addresses = addresses
	for _, address := range addresses {
		c := context.Background()
		c, _ = context.WithTimeout(c, time.Second*30)
		conn, err := grpc.DialContext(c, address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		d.conns = append(d.conns, conn)
	}
	// This generates a unique folder structure for maintaining client cache
	d.client = client.NewDgraphClient(d.conns, client.DefaultOptions, clientDir)
	return d, err
}

// This function creates Dgraph object with provided client
// No alterations are made, neigher any checks
func ConnectWithClient(c *client.Dgraph) (*Dgraph, error) {
	return &Dgraph{client: c}, nil
}
