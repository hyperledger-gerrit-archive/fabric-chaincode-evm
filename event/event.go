/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

type MessageInfo struct {
	ctx     context.Context
	message interface{}
	tags    map[string]interface{}
}

type MessagePayload struct {
	Message interface{}
}

type EventManager struct {
	stub       shim.ChaincodeStubInterface
	eventCache []MessageInfo
	publisher  event.Publisher
}

func NewEventManager(stub shim.ChaincodeStubInterface, publisher event.Publisher) *EventManager {
	return &EventManager{
		stub:       stub,
		eventCache: make([]MessageInfo, 0),
		publisher:  publisher,
	}
}

func (evmgr *EventManager) Flush() error {
	var err error
	var eventMsgs []MessagePayload
	eventMsgs = make([]MessagePayload, 0)

	for i := 0; i < len(evmgr.eventCache); i++ {
		msg := MessagePayload{Message: evmgr.eventCache[i].message}
		eventMsgs = append(eventMsgs, msg)
	}

	payload, ok := json.Marshal(eventMsgs)
	if ok != nil {
		return fmt.Errorf("error in marshaling event messages: %s", ok.Error())
	}
	err = evmgr.stub.SetEvent("Chaincode event", payload)
	return err
}

func (evmgr *EventManager) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	evmgr.eventCache = append(evmgr.eventCache, MessageInfo{
		ctx:     ctx,
		message: message,
		tags:    tags,
	})
	return nil
}
