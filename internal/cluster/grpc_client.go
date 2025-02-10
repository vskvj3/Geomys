package cluster

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	pb "github.com/vskvj3/geomys/internal/cluster/proto" // Import generated gRPC code

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

type GrpcClient struct {
	Conn          *grpc.ClientConn
	Client        pb.NodeServiceClient
	LeaderID      int
	LastHeartbeat time.Time
	Nodes         map[int32]string
}

// Connect to a node's gRPC server
func NewGrpcClient(address string) (*GrpcClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	client := pb.NewNodeServiceClient(conn)
	return &GrpcClient{Conn: conn, Client: client}, nil
}

// Request vote from another node
func (c *GrpcClient) RequestVote(nodeID int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.Client.RequestVote(ctx, &pb.VoteRequest{NodeId: int32(nodeID)})
	if err != nil {
		log.Printf("Vote request failed: %v", err)
		return false
	}
	return resp.Granted
}

// Send heartbeat to leader
func (c *GrpcClient) SendHeartbeat(nodeID int, port int32) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// this is for debug, remove after use
	peerInfo, ok := peer.FromContext(ctx)
	var addr string
	if ok {
		addr = peerInfo.Addr.String()
		fmt.Printf("sending heartbeat to %s\n", addr)
	}
	/***/
	resp, err := c.Client.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: int32(nodeID), Port: strconv.Itoa(int(port))})
	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
		return false
	}
	c.LastHeartbeat = time.Now()
	fmt.Println("grpc_client.go: sendHeartBeat", resp.Nodes)
	c.Nodes = resp.Nodes
	return resp.Success
}

func (c *GrpcClient) MonitorLeader(server *GrpcServer) {
	for {
		success := c.SendHeartbeat(int(server.NodeID), server.Port)
		if !success {
			fmt.Println("Failed to send heartbeat. Checking leader status...")

			if time.Since(c.LastHeartbeat) > 15*time.Second {
				fmt.Println("Leader appears to be down. Starting leader election...")

				c.StartLeaderElection(server)

				fmt.Println(c.LeaderID, int(server.NodeID))

				if c.LeaderID == int(server.NodeID) {
					return
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// Initiate leader election
func (c *GrpcClient) StartLeaderElection(s *GrpcServer) {
	// Determine new leader (lowest ID)
	newLeader := s.NodeID
	fmt.Println("Checking for new leaders")
	for nodeId := range c.Nodes {
		fmt.Println("current node id: ", nodeId, " prev newLeaderid", newLeader)
		if int(nodeId) < int(newLeader) {
			newLeader = int32(nodeId)
		}
	}
	fmt.Println("THIS IS FROM THE LEADER ELECTION")
	fmt.Println("server id:", s.NodeID)
	fmt.Println("nodes", c.Nodes)
	fmt.Println("new leader: ", newLeader)
	fmt.Println("newLeader adress: ", c.Nodes[newLeader])

	// If this node is the new leader, announce it
	if newLeader == s.NodeID {
		fmt.Printf("Inside grpc server: Node %d is now the new leader\n", s.NodeID)
		c.LeaderID = int(s.NodeID)
		go s.MonitorFollowers()
	} else {
		fmt.Printf("Inside grpc server: Node %d is elected as the new leader\n", newLeader)
		c.LeaderID = int(newLeader)
		// conn, err := grpc.NewClient(c.Nodes[newLeader], grpc.WithTransportCredentials(insecure.NewCredentials()))
		conn, err := grpc.NewClient("127.0.0.1:2010", grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			fmt.Println("New connection failed")
		}
		c.Conn = conn
		c.Client = pb.NewNodeServiceClient(conn)
	}
}
