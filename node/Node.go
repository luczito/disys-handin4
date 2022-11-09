package main

import (
	api "github.com/hashicorp/consul/api"
	service "github.com/luczito/disys-handin4/proto"
)

type node struct {
	Name    string
	Address string

	sAddress string
	sdkv     api.KV

	Clients map[string]service.Serviceclient
}
