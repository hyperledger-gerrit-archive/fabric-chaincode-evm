/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

type FabProxy struct {
	rpcServer  *rpc.Server
	httpServer *http.Server
}

func NewFabProxy(service EthService) (*FabProxy, error) {
	rpcServer := rpc.NewServer()

	proxy := &FabProxy{
		rpcServer: rpcServer,
	}

	rpcServer.RegisterCodec(NewRPCCodec(), "application/json")
	if err := rpcServer.RegisterService(service, "eth"); err != nil {
		return nil, err
	}
	if err := rpcServer.RegisterService(&NetService{}, "net"); err != nil {
		return nil, err
	}
	return proxy, nil
}

func (p *FabProxy) Start(port int) error {
	r := mux.NewRouter()
	r.Handle("/", p.rpcServer)

	allowedHeaders := handlers.AllowedHeaders([]string{"Origin", "Content-Type"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"POST"})

	p.httpServer = &http.Server{Handler: handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(r), Addr: fmt.Sprintf(":%d", port)}
	return p.httpServer.ListenAndServe()
}

func (p *FabProxy) Shutdown() error {
	return p.httpServer.Shutdown(context.Background())
}
