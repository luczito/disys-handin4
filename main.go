package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	token "github.com/luczito/disys-handin4/grpc"
	"google.golang.org/grpc"
)

var port = flag.Int("port", 5000, "port") //port for the node default 5000

type STATE int32 //state struct to see wether the node has access, wants access or not.

const (
	RELEASED STATE = iota
	HELD
	REQUESTED
)

// node struct.
type node struct {
	token.UnimplementedRingServer
	id       int32 //port
	clients  map[int32]token.RingClient
	ctx      context.Context
	queue    []int32
	state    STATE
	requests int
}

// request function to request access to the critical service
func (n *node) RequestAccess(ctx context.Context, req *token.Request) (*token.Ack, error) {
	log.Printf("Request from %v recieved.", req.Id)

	//checks if state is higher, if equal check id. (This is the hierarchy)
	if n.state == HELD || (n.state == REQUESTED && (n.id > req.Id)) {
		log.Printf("Queueing request from %v\n", req.Id)
		n.queue = append(n.queue, req.Id)
	} else {
		if n.state == REQUESTED {
			n.requests++
			n.clients[req.Id].RequestAccess(ctx, &token.Request{Id: n.id})
		}
		log.Printf("Sending reply to %v", req.Id)
		n.clients[req.Id].Reply(ctx, &token.Reply{})
	}
	reply := &token.Ack{}
	return reply, nil
}

// reply function to check wether all nodes have replied to the request.
func (n *node) Reply(ctx context.Context, req *token.Reply) (*token.AckReply, error) {
	n.requests--

	if n.requests == 0 {
		log.Println("Recieved reply. All requests replied.")
		go n.CriticalSection()
	}

	log.Printf("Recieved reply. Missing %d replies.\n", n.requests)

	rep := &token.AckReply{}
	return rep, nil
}

// critical section emulated in this method. takes 5 seconds to exec
func (n *node) CriticalSection() {
	log.Println("Critical Section accessed")

	n.state = HELD

	time.Sleep(5 * time.Second)

	n.state = RELEASED

	log.Println("Critical Section done, releasing")

	n.ReplyQueue()
}

// Requests all other nodes.
func (n *node) sendRequestToAll() {
	n.state = REQUESTED

	request := &token.Request{
		Id: n.id,
	}

	n.requests = len(n.clients)

	log.Printf("Sending request to all other nodes. Missing %d replies.\n", n.requests)

	for id, client := range n.clients {
		_, err := client.RequestAccess(n.ctx, request)

		if err != nil {
			log.Printf("Something went wrong with node: %v, error: %v", id, err)
		}
	}
}

// Sends replies to all requests in the queue, then emptied the queue
func (n *node) ReplyQueue() {
	reply := &token.Reply{}

	for _, id := range n.queue {
		_, err := n.clients[id].Reply(n.ctx, reply)

		if err != nil {
			log.Printf("Something went wrong\n", id)
		}
	}
	n.queue = make([]int32, 0)
}

func main() {
	flag.Parse() //set port with -port in the commandline when running the program

	ctx_, cancel := context.WithCancel(context.Background())
	defer cancel()

	var port = int32(*port)

	//create a node for this proccess
	n := &node{
		id:       port,
		clients:  make(map[int32]token.RingClient),
		queue:    make([]int32, 0),
		ctx:      ctx_,
		state:    RELEASED,
		requests: 0,
	}

	//creates listener on port
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}

	log.Printf("Node created on port: %d\n", n.id)

	grpcServer := grpc.NewServer()
	token.RegisterRingServer(grpcServer, n)

	//serve on the listener
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	//for loop to dial all other nodes in the network, if this loop is increased the number of nodes in the network is aswell
	for i := 0; i < 3; i++ {
		nodePort := int32(5000 + i)

		if nodePort == n.id {
			continue
		}

		var conn *grpc.ClientConn
		log.Printf("Trying to dial: %v\n", nodePort)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", nodePort), grpc.WithInsecure(), grpc.WithBlock())

		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}

		defer conn.Close()
		log.Printf("Succes connecting to: %v\n", nodePort)
		c := token.NewRingClient(conn)
		n.clients[nodePort] = c
	}

	log.Printf("I am connected to %v other nodes", len(n.clients))

	scanner := bufio.NewScanner(os.Stdin)

	//scanner that requests access from the given node when something is written in the terminal.
	for scanner.Scan() {
		log.Println("Requesting access to critical section...")
		n.sendRequestToAll()
	}
}
