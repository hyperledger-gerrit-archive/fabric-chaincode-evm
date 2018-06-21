/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"context"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// When exceeded, we will trim the buffer's backing array capacity to avoid excessive allocation
const maximumBufferCapacityToLengthRatio = 2

type Publisher interface {
	Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error
}

type Cache struct {
	publisher Publisher
	events    []messageInfo
}

type messageInfo struct {
	ctx     context.Context
	message interface{}
	tags    map[string]interface{}
}

type eventManager struct {
	stub       shim.ChaincodeStubInterface
	eventCache Cache
}

func NewEventManager(stub shim.ChaincodeStubInterface, publisher Publisher) eventManager {
	return eventManager{
		stub: stub,
		eventCache: Cache{
			publisher: publisher,
		},
	}
}

func (evmgr *eventManager) Flush() error {
	var err error
	for _, mi := range evmgr.eventCache.events {
		publishErr := evmgr.Publish(mi.ctx, mi.message, mi.tags)
		// Capture first by try to flush the rest
		if publishErr != nil && err == nil {
			err = publishErr
		}
	}
	// Clear the buffer by re-slicing its length to zero
	if cap(evmgr.eventCache.events) > len(evmgr.eventCache.events)*maximumBufferCapacityToLengthRatio {
		// Trim the backing array capacity when it is more than double the length of the slice to
		//avoid tying up memory after a spike in the number of events to buffer
		evmgr.eventCache.events = evmgr.eventCache.events[:0:len(evmgr.eventCache.events)]
	} else {
		// Re-slice the length to 0 to clear buffer but hang on to spare capacity in backing array
		//that has been added in previous cache round
		evmgr.eventCache.events = evmgr.eventCache.events[:0]
	}
	return err
}

func (evmgr *eventManager) Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error {
	//return em.pubsubServer.PublishWithTags(ctx, message, tagMap(tags))

	//The message here was an EventDataLog object
	//So, step 1: marshal the message into an EventDataLog type object, evData
	//var evData events.EventDataLog
	//Then,
	//err := evmgr.stub.SetEvent(evData.Topics[0].String(), []byte("Event is set"))
	err := evmgr.stub.SetEvent("Chaincode event", []byte("Event is set"))
	return err
}
