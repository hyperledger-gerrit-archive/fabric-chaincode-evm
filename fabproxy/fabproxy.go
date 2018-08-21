/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
)

type FabProxy struct {
	server *rpc.Server
}

func NewFabProxy(service EthService) *FabProxy {
	server := rpc.NewServer()

	proxy := &FabProxy{
		server: server,
	}

	server.RegisterCodec(NewRPCCodec(), "application/json")
	server.RegisterService(service, "eth")

	return proxy
}

func (p *FabProxy) Start(port int) {
	r := mux.NewRouter()
	r.Handle("/", p.server)

	allowedHeaders := handlers.AllowedHeaders([]string{"Origin", "Content-Type"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"POST"})

	http.ListenAndServe(fmt.Sprintf(":%d", port), handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(r))
}
