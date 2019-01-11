/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"

	"github.com/hyperledger/fabric-chaincode-evm/fab3"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const usage = `Fab3 uses environment variables to be able to start communicating with a Fabric network
	Required Environment Variables:
	  FAB3_CONFIG - Path to a compatible Fabric SDK Go config file
	  FAB3_USER - User identity being used for the proxy (Matches the users names in the crypto-config directory specified in the config)
	  FAB3_ORG - Organization of the specified user
	  FAB3_CHANNEL - Channel to be used for the transactions
	  FAB3_CCID - ID of the EVM Chaincode deployed in your fabric network

	Other Environment Variables:
	  PORT - Port the Fab3 will be running on. Default is 5000
	`

var logger *zap.SugaredLogger

func main() {
	rawLogger, _ := zap.NewProduction()
	logger = rawLogger.Named("fab3").Sugar()

	cfg := grabEnvVar("FAB3_CONFIG", true)
	org := grabEnvVar("FAB3_ORG", true)
	user := grabEnvVar("FAB3_USER", true)
	ch := grabEnvVar("FAB3_CHANNEL", true)
	ccid := grabEnvVar("FAB3_CCID", true)
	port := grabEnvVar("PORT", false)

	portNumber := 5000
	if port != "" {
		var err error
		portNumber, err = strconv.Atoi(port)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to convert the environment variable `PORT`, %s,  to an int\n", port)
			os.Exit(1)
		}
	}

	sdk, err := fabsdk.New(config.FromFile(cfg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Fabric SDK Client: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	clientChannelContext := sdk.ChannelContext(ch, fabsdk.WithUser(user), fabsdk.WithOrg(org))
	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Fabric SDK Channel Client: %s\n", err)
		os.Exit(1)
	}

	ledger, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Fabric SDK Ledger Client: %s\n", err)
		os.Exit(1)
	}

	ethService := fab3.NewEthService(client, ledger, ch, ccid, logger)

	logger.Infof("Starting Fab3 on port %d\n", portNumber)
	proxy := fab3.NewFab3(ethService)
	err = proxy.Start(portNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Fab3: %s", err)
	}
	defer func() {
		logger.Info("Shutting down Fab3")
		proxy.Shutdown()
		logger.Info("Fab3 has exited")
	}()
}

func grabEnvVar(varName string, required bool) string {
	envVar := os.Getenv(varName)
	if required && envVar == "" {
		fmt.Fprintf(os.Stderr, "Fab3 requires the environment variable %s to be set\n\n%s\n\n", varName, usage)
		os.Exit(1)
	}
	return envVar
}
