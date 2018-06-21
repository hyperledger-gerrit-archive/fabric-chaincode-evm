/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"context"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution/evm/events"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric/core/chaincode/shim"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {

	var (
		em            eventManager
		mockStub      *mocks.MockStub
		publisher     Publisher
		addr          account.Address
		fakeGetLedger map[string][]byte
		fakePutLedger map[string][]byte
	)

	BeforeEach(func() {
		mockStub = &mocks.MockStub{}
		em = NewEventManager(mockStub, publisher)

		var err error
		addr, err = account.AddressFromBytes([]byte("0000000000000address"))
		Expect(err).ToNot(HaveOccurred())
		fakeGetLedger = make(map[string][]byte)
		fakePutLedger = make(map[string][]byte)

		//Writing to a separate ledger so that writes to the ledger cannot be read
		//in the same transaction. This is more consistent with the behavior of
		//the ledger
		mockStub.PutStateStub = func(key string, value []byte) error {
			fakePutLedger[key] = value
			return nil
		}

		mockStub.GetStateStub = func(key string) ([]byte, error) {
			return fakeGetLedger[key], nil
		}

		mockStub.DelStateStub = func(key string) error {
			delete(fakePutLedger, key)
			return nil
		}
	})

	Describe("Publish", func() {
		var (
			ctx     context.Context
			message interface{}
			tags    map[string]interface{}
		)

		BeforeEach(func() {
			ctx = context.Background()
			message = events.EventDataLog{
				Address: addr,
				Height:  0,
			}
			tags = make(map[string]interface{})
		})

		Context("when Publish() is called", func() {
			It("sets the chaincode event", func() {
				err := em.Publish(ctx, message, tags)
				Expect(err).ToNot(HaveOccurred())
				n, p := mockStub.SetEventArgsForCall(0)
				Expect(n).To(Equal("Chaincode event"))
				Expect(p).To(Equal([]byte("Event is set")))
			})
		})
	})
})

type Publisher interface {
	Publish(ctx context.Context, message interface{}, tags map[string]interface{}) error
}

type eventManager struct {
	stub       shim.ChaincodeStubInterface
	eventCache Cache
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

func NewEventManager(stub shim.ChaincodeStubInterface, publisher Publisher) eventManager {
	return eventManager{
		stub: stub,
		eventCache: Cache{
			publisher: publisher,
		},
	}
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
