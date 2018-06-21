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
	Ctx     context.Context        `json:"ctx"`
	Message interface{}            `json:"message"`
	Tags    map[string]interface{} `json:"tags"`
}

type MessagePayload struct {
	Message interface{}
}

type EventManager struct {
	stub       shim.ChaincodeStubInterface
	EventCache []MessageInfo
	publisher  event.Publisher
}

func NewEventManager(stub shim.ChaincodeStubInterface, publisher event.Publisher) *EventManager {
	return &EventManager{
		stub:       stub,
		EventCache: make([]MessageInfo, 0),
		publisher:  publisher,
	}
}

func (evmgr *EventManager) Flush(eventName string) error {
	var err error
	var eventMsgs []MessagePayload
	eventMsgs = make([]MessagePayload, 0)

	if len(evmgr.EventCache) > 0 {
		for i := 0; i < len(evmgr.EventCache); i++ {
			msg := MessagePayload{Message: evmgr.EventCache[i].Message}
			eventMsgs = append(eventMsgs, msg)
		}

		payload, ok := json.Marshal(eventMsgs)
		if ok != nil {
			return fmt.Errorf("error in marshaling event messages: %s", ok.Error())
		}
		err = evmgr.stub.SetEvent(eventName, payload)
		return err
	}

	return nil
}

func (evmgr *EventManager) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	evID, ok := tags["EventID"].(string)
	if !ok {
		return fmt.Errorf("type mismatch: expected string but received %T", tags["EventID"])
	}

	//Burrow EVM emits other events related to state (such as account call) as well, but we are only interested in log events
	if evID[0:3] == "Log" {
		evmgr.EventCache = append(evmgr.EventCache, MessageInfo{
			Ctx:     ctx,
			Message: message,
			Tags:    tags,
		})
	}
	return nil
}
