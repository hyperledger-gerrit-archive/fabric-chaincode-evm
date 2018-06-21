/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"context"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type messageInfo struct {
	ctx     context.Context
	message interface{}
	tags    map[string]interface{}
}

type EventManager struct {
	stub       shim.ChaincodeStubInterface
	eventCache []messageInfo
}

func NewEventManager(stub shim.ChaincodeStubInterface) EventManager {
	return EventManager{
		stub:       stub,
		eventCache: make([]messageInfo, 0),
	}
}

func (evmgr *EventManager) Flush() error {
	var err error
	var payload []byte
	payload = make([]byte, 0)
	for _, mi := range evmgr.eventCache {
		m, ok := mi.message.(byte)
		if ok {
			payload = append(payload, m)
		}
	}
	evmgr.stub.SetEvent("Chaincode event", payload)
	return err
}

func (evmgr *EventManager) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	evmgr.eventCache = append(evmgr.eventCache, messageInfo{
		ctx:     ctx,
		message: message,
		tags:    tags,
	})
	return nil
}
