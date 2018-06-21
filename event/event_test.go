/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm/events"
	evm_event "github.com/hyperledger/fabric-chaincode-evm/event"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event", func() {

	var (
		em        evm_event.EventManager
		mockStub  *mocks.MockStub
		addr      account.Address
		publisher event.Publisher
	)

	BeforeEach(func() {
		mockStub = &mocks.MockStub{}
		em = *evm_event.NewEventManager(mockStub, publisher)

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
			tags = map[string]interface{}{"EventID": fmt.Sprintf("Log/%s", addr)}
		})

		Context("when an event is emitted by calling the publish function", func() {
			Context("if it is a log event", func() {
				It("appends the new message info into the eventCache", func() {
					l1 := len(em.EventCache)
					err := em.Publish(ctx, message, tags)
					Expect(err).ToNot(HaveOccurred())
					l2 := len(em.EventCache)
					Expect(l2).To(Equal(l1 + 1))
					mi := evm_event.MessageInfo{
						Ctx:     ctx,
						Message: message,
						Tags:    tags,
					}
					Expect(em.EventCache[l2-1]).To(Equal(mi))
				})
			})

			Context("if it is not a log event", func() {
				It("does nothing", func() {
					l1 := len(em.EventCache)
					e1 := em.EventCache
					var alt_tags map[string]interface{}
					alt_tags = map[string]interface{}{"EventID": fmt.Sprintf("Acc/%s/Call", addr)}
					err := em.Publish(ctx, message, alt_tags)
					Expect(err).ToNot(HaveOccurred())
					l2 := len(em.EventCache)
					e2 := em.EventCache
					Expect(l2).To(Equal(l1))
					Expect(e2).To(Equal(e1))
				})
			})
		})

		Context("when an error occurs", func() {
			It("is due to type mismatch", func() {
				var err_tags map[string]interface{}
				err_tags = map[string]interface{}{"EventID": []byte(fmt.Sprintf("Log/%s", addr))}
				err := em.Publish(ctx, message, err_tags)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Flush", func() {
		var (
			ctx      context.Context
			message1 interface{}
			message2 interface{}
			tags     map[string]interface{}
			payload1 []byte
			payload2 []byte
			mp1      []evm_event.MessagePayload
			mp2      []evm_event.MessagePayload
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
			tags = map[string]interface{}{"EventID": fmt.Sprintf("Log/%s", addr)}
			mp1 = make([]evm_event.MessagePayload, 0)
			mp2 = make([]evm_event.MessagePayload, 0)
			mp1 = append(mp1, evm_event.MessagePayload{Message: message1})
			var ok1 error
			var ok2 error
			payload1, ok1 = json.Marshal(mp1)
			Expect(ok1).ToNot(HaveOccurred())
			mp2 = append(mp2, evm_event.MessagePayload{Message: message1})
			mp2 = append(mp2, evm_event.MessagePayload{Message: message2})
			payload2, ok2 = json.Marshal(mp2)
			Expect(ok2).ToNot(HaveOccurred())
		})

		Context("when a single event is emitted", func() {
			It("sets a new event with a single messageInfo object payload", func() {
				err := em.Publish(ctx, message1, tags)
				Expect(err).ToNot(HaveOccurred())
				er := em.Flush("Chaincode event")
				Expect(er).ToNot(HaveOccurred())
				Expect(mockStub.SetEventCallCount()).To(Equal(1))
				n, p := mockStub.SetEventArgsForCall(0)
				Expect(n).To(Equal("Chaincode event"))
				Expect(p).To(Equal(payload1))

				var a []struct{ Message events.EventDataLog }
				e := json.Unmarshal(p, &a)
				Expect(e).ToNot(HaveOccurred())
				Expect(a[0].Message).To(Equal(message1))
			})
		})

		Context("when multiple events are emitted", func() {
			It("sets a new event with a payload consisting of messageInfo objects marshaled together", func() {
				err := em.Publish(ctx, message1, tags)
				Expect(err).ToNot(HaveOccurred())
				err1 := em.Publish(ctx, message2, tags)
				Expect(err1).ToNot(HaveOccurred())
				er := em.Flush("Chaincode event")
				Expect(er).ToNot(HaveOccurred())
				Expect(mockStub.SetEventCallCount()).To(Equal(1))
				n, p := mockStub.SetEventArgsForCall(0)
				Expect(n).To(Equal("Chaincode event"))
				Expect(p).To(Equal(payload2))

				var a []struct{ Message events.EventDataLog }
				e := json.Unmarshal(p, &a)
				Expect(e).ToNot(HaveOccurred())
				Expect(a[0].Message).To(Equal(message1))
				Expect(a[1].Message).To(Equal(message2))
			})
		})

		Context("when an error occurs", func() {
			Context("due to problems in marshaling event messages", func() {
				It("returns an error", func() {
					msg1 := make(chan events.EventDataLog)
					err := em.Publish(ctx, msg1, tags)
					Expect(err).ToNot(HaveOccurred())
					er := em.Flush("Chaincode event")
					Expect(er).To(HaveOccurred())
				})
			})

			Context("due to invalid event name (nil string)", func() {
				BeforeEach(func() {
					mockStub.SetEventReturns(errors.New("error: nil event name"))
				})

				It("returns an error", func() {
					err := em.Publish(ctx, message1, tags)
					Expect(err).ToNot(HaveOccurred())
					err1 := em.Publish(ctx, message2, tags)
					Expect(err1).ToNot(HaveOccurred())
					er := em.Flush("")
					Expect(er).To(HaveOccurred())
				})
			})
		})
	})
})
