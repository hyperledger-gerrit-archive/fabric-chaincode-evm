/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func main() {
	cfg := grabEnvVar("FABPROXY_CONFIG", true, "Path to the Fabric SDK GO config file")
	org := grabEnvVar("FABPROXY_ORG", true, "Org of the user specified")
	user := grabEnvVar("FABPROXY_USER", true, "User to be used for proxy")
	ch := grabEnvVar("FABPROXY_CHANNEL", true, "Channel transactions will be sent on")
	ccid := grabEnvVar("FABPROXY_CCID", true, "Chaincode ID of the EVM chaincode")
	port := grabEnvVar("PORT", false, "")

	var portNumber int
	if port != "" {
		var err error
		portNumber, err = strconv.Atoi(port)
		if err != nil {
			fmt.Printf("Failed to convert the environment variable `PORT`, %s,  to an int\n", port)
			os.Exit(1)
		}
	} else {
		portNumber = 5000
	}

	sdk, err := fabsdk.New(config.FromFile(cfg))
	if err != nil {
		fmt.Printf("Failed to create Fabric SDK Client: %s\n", err.Error())
		os.Exit(1)
	}
	defer sdk.Close()

	clientChannelContext := sdk.ChannelContext(ch, fabsdk.WithUser(user), fabsdk.WithOrg(org))
	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create Fabric SDK Channel Client: %s\n", err.Error())
		os.Exit(1)
	}

	ledger, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create Fabric SDK Ledger Client: %s\n", err.Error())
		os.Exit(1)
	}

	ethService := fabproxy.NewEthService(client, ledger, ch, ccid)

	fmt.Printf("Starting Fab Proxy on port %d\n", portNumber)
	proxy := fabproxy.NewFabProxy(ethService)
	proxy.Start(portNumber)
}

func grabEnvVar(varName string, errorIfEmpty bool, description string) string {
	envVar := os.Getenv(varName)
	if errorIfEmpty && envVar == "" {
		fmt.Printf("Fab Proxy requires the environment variable %s to be set\n", varName)
		fmt.Printf("Description of %s: %s\n", varName, description)
		os.Exit(1)
	}
	return envVar
}
