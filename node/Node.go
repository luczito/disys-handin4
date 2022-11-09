package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	api "github.com/hashicorp/consul/api"
	service "github.com/luczito/disys-handin4/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type node struct {
	Name    string
	Address string

	ConsulAddress string
	sdkv          api.KV

	Clients map[string]service.Serviceclient
}

func (n *node) SayHello(ctx context.Context, in *service.ConnectRequest) (*service.ConnectReply, error) {
	return &service.ConnectReply{Message: "Connection from " + n.Name}, nil
}

func (n *node) StartListener() {
	listener, err := net.Listen("tcp", n.Address)
	if err != nil {
		log.Fatalf("Unable to listen on tcp %v: error: %v", n.Address, err)
	}

	n_ := grpc.NewServer()

	service.RegisterServiceService(n_, n)

	reflection.Register(n_)

	err_ := n_.Serve(listener)
	if err != nil {
		log.Fatalf("unable to serve server %v", err_)
	}
}

func (n *node) Register() {
	config := api.DefaultConfig()

	config.Address = n.ConsulAddress

	consul, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Unable to reach service discovery %v", err)
	}

	kv := consul.KV()
	peer := &api.KVPair{Key: n.Name, Value: []byte(n.Address)}
	_, err = kv.Put(peer, nil)
	if err != nil {
		log.Fatalf("Unable to register with service discovery")
	}

	n.sdkv = *kv

	log.Printf("Successfully registered with service discovery")
}

func (n *node) Start() {
	n.Clients = make(map[string]service.ServiceClient)

	go n.StartListener()

	n.Register()

	for {
		time.Sleep(20 * time.Second)
		n.FindAll()
	}
}

func (n *node) SetupClient(name string, address string) {
	connection, err := grpc.Dial(address, grpc.WithInsecure)
	if err != nil {
		log.Fatalf("Could not connect %v", err)
	}
	defer connection.Close()

	n.Clients[name] = service.NewServiceClient(connection)

	reply, err := n.Clients[name].Greet(context.Background(), &service.ConnectRequest{Name: n.Name})
	if err != nil {
		log.Fatalf("Unable to connect to other nodes %v", err)
	}
	log.Printf("Reply from other node: %v", reply)
}

func (n *node) FindAll() {
	kvpairs, _, err := n.sdkv.List("Node", nil)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	for _, kventry := range kvpairs {
		if strings.Compare(kventry.Key, n.Name) == 0 {
			continue
		}
		if n.Clients[kventry.Key] == "" {
			log.Printf("New node: %v", kventry.Key)
			n.SetupClient(kventry.Key, string(kventry.Value))
		}
	}
}

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		fmt.Println("arguments required: name, listening address, consul address")
		os.Exit(1)
	}

	name := args[0]
	listenAddress := args[1]
	consulAddress := args[2]

	node_ := node{Name: name, Address: listenAddress, ConsulAddress: consulAddress, Clients: nil}

	node_.Start()
}
