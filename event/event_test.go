/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"context"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/execution/evm/events"
	evm_event "github.com/hyperledger/fabric-chaincode-evm/event"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {

	var (
		em       evm_event.EventManager
		mockStub *mocks.MockStub
		addr     account.Address
	)

	BeforeEach(func() {
		mockStub = &mocks.MockStub{}
		em = evm_event.NewEventManager(mockStub)

		var err error
		addr, err = account.AddressFromBytes([]byte("0000000000000address"))
		Expect(err).ToNot(HaveOccurred())
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
			It("appends the new message info into the eventCache", func() {
				err := em.Publish(ctx, message, tags)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("Flush", func() {
		var (
			ctx      context.Context
			message1 interface{}
			message2 interface{}
			tags     map[string]interface{}
			payload  []byte
		)

		BeforeEach(func() {
			ctx = context.Background()
			message1 = events.EventDataLog{
				Address: addr,
				Height:  0,
			}
			message2 = events.EventDataLog{
				Address: addr,
				Height:  1,
			}
			tags = make(map[string]interface{})

			payload = make([]byte, 0)
		})

		Context("when Flush() is called", func() {
			Context("when a single event is emitted", func() {
				It("sets a new event with a single messageInfo object payload", func() {
					err := em.Publish(ctx, message1, tags)
					Expect(err).ToNot(HaveOccurred())
					m, ok := message1.(byte)
					if ok {
						payload = append(payload, m)
					}
					em.Flush()
					n, p := mockStub.SetEventArgsForCall(0)
					Expect(n).To(Equal("Chaincode event"))
					Expect(p).To(Equal(payload))
					Expect(mockStub.SetEventCallCount()).To(Equal(1))
				})
			})

			Context("when multiple events are emitted", func() {
				It("sets a new event with a payload consisting of messageInfo objects put together", func() {
					err := em.Publish(ctx, message1, tags)
					Expect(err).ToNot(HaveOccurred())
					err1 := em.Publish(ctx, message2, tags)
					Expect(err1).ToNot(HaveOccurred())
					m, ok := message1.(byte)
					if ok {
						payload = append(payload, m)
					}
					m, ok = message2.(byte)
					if ok {
						payload = append(payload, m)
					}
					em.Flush()
					n, p := mockStub.SetEventArgsForCall(0)
					Expect(n).To(Equal("Chaincode event"))
					Expect(p).To(Equal(payload))
					Expect(mockStub.SetEventCallCount()).To(Equal(1))
				})
			})
		})
	})
})
