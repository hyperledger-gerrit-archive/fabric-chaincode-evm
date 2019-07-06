/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab3

import (
	"net/http"
)

// NetworkID is the hex encoding of the string "fabric-evm"
const NetworkID = "0x6661627269632d65766d"

// NetService returns data about the network the client is connected
// to.
type NetService struct {
}

// Version takes no parameters and returns the network identifier.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#net_version
func (s *NetService) Version(r *http.Request, _ *interface{}, reply *string) error {
	*reply = NetworkID
	return nil
}
