package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	token "github.com/luczito/disys-handin4/grpc"
	"google.golang.org/grpc"
)

type node struct {
	token.UnimplementedRingServer
	id         int32 //port
	clients    map[int32]token.RingClient
	ctx        context.Context
	hasToken   bool
	neighbor   *node
	requestAcc bool
}

func main() {
	//set port for node in commandline
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	arg2, _ := strconv.ParseBool(os.Args[2])
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := &node{
		id:         ownPort,
		clients:    make(map[int32]token.RingClient),
		ctx:        ctx,
		hasToken:   arg2,
		requestAcc: false,
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	//register server
	grpcServer := grpc.NewServer()
	token.RegisterRingServer(grpcServer, n)

	//serve server
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	//number of nodes in the system
	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}
		//dial to other nodes
		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := token.NewRingClient(conn)
		n.clients[port] = c
	}
	n.tokenCheck()
}

func (n *node) requestAccess() {
	n.requestAcc = true
}

func (n *node) tokenCheck() {
	for {
		if n.hasToken == true {
			log.Printf("%v: has the token", n.id)
			if n.requestAcc != true {
				log.Printf("%v: is passing the token on", n.id)
				n.neighbor.hasToken = true
				n.hasToken = false
			} else {
				n.criticalService()
			}
		}
	}
}

func (n *node) criticalService() {
	log.Printf("%v: has access to the critical server", n.id)
	fmt.Println("Critical service accessed by %v", n.id)
}
