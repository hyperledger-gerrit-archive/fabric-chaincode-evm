/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/execution/evm/events"
	evm_event "github.com/hyperledger/fabric-chaincode-evm/event"
	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var evmcc = "evmcc"
var _ = Describe("Ethservice", func() {
	var (
		ethservice fabproxy.EthService

		mockChClient     *mocks.MockChannelClient
		mockLedgerClient *mocks.MockLedgerClient
		channelID        string
	)

	BeforeEach(func() {
		mockChClient = &mocks.MockChannelClient{}
		mockLedgerClient = &mocks.MockLedgerClient{}
		channelID = "test-channel"

		ethservice = fabproxy.NewEthService(mockChClient, mockLedgerClient, channelID, evmcc)
	})

	Describe("GetCode", func() {
		var (
			sampleCode    []byte
			sampleAddress string
		)

		BeforeEach(func() {
			sampleCode = []byte("sample-code")
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleCode,
			}, nil)

			sampleAddress = "1234567123"
		})

		It("returns the code associated to that address", func() {
			var reply string

			err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: evmcc,
				Fcn:         "getCode",
				Args:        [][]byte{[]byte(sampleAddress)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleCode)))
		})

		Context("when the address has `0x` prefix", func() {
			BeforeEach(func() {
				sampleAddress = "0x123456"
			})
			It("returns the code associated with that address", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.QueryCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.QueryArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         "getCode",
					Args:        [][]byte{[]byte(sampleAddress[2:])},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleCode)))
			})
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})

	Describe("Call", func() {
		var (
			encodedResponse []byte
			sampleArgs      *fabproxy.EthArgs
		)

		BeforeEach(func() {
			sampleResponse := []byte("sample response")
			encodedResponse = make([]byte, hex.EncodedLen(len(sampleResponse)))
			hex.Encode(encodedResponse, sampleResponse)
			mockChClient.QueryReturns(channel.Response{
				Payload: sampleResponse,
			}, nil)

			sampleArgs = &fabproxy.EthArgs{
				To:   "1234567123",
				Data: "sample-data",
			}
		})

		It("returns the value of the simulation of executing a smart contract with a `0x` prefix", func() {

			var reply string

			err := ethservice.Call(&http.Request{}, sampleArgs, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: evmcc,
				Fcn:         sampleArgs.To,
				Args:        [][]byte{[]byte(sampleArgs.Data)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal("0x" + string(encodedResponse)))
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when the address has a `0x` prefix", func() {
			BeforeEach(func() {
				sampleArgs.To = "0x" + sampleArgs.To
			})
			It("strips the prefix from the query", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, sampleArgs, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.QueryCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.QueryArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         sampleArgs.To[2:],
					Args:        [][]byte{[]byte(sampleArgs.Data)},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal("0x" + string(encodedResponse)))
			})
		})

		Context("when the data has a `0x` prefix", func() {
			BeforeEach(func() {
				sampleArgs.Data = "0x" + sampleArgs.Data
			})

			It("strips the prefix from the query", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, sampleArgs, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.QueryCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.QueryArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         sampleArgs.To,
					Args:        [][]byte{[]byte(sampleArgs.Data[2:])},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal("0x" + string(encodedResponse)))
			})
		})
	})

	Describe("SendTransaction", func() {
		var (
			sampleResponse channel.Response
			sampleArgs     *fabproxy.EthArgs
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{
				Payload:       []byte("sample-response"),
				TransactionID: "1",
			}
			mockChClient.ExecuteReturns(sampleResponse, nil)

			sampleArgs = &fabproxy.EthArgs{
				To:   "1234567123",
				Data: "sample-data",
			}
		})

		It("returns the transaction id", func() {
			var reply string
			err := ethservice.SendTransaction(&http.Request{}, sampleArgs, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.ExecuteCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.ExecuteArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: evmcc,
				Fcn:         sampleArgs.To,
				Args:        [][]byte{[]byte(sampleArgs.Data)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
		})

		Context("when the transaction is a contract deployment", func() {
			BeforeEach(func() {
				sampleArgs.To = ""
			})

			It("returns the transaction id", func() {
				var reply string
				err := ethservice.SendTransaction(&http.Request{}, sampleArgs, &reply)
				Expect(err).ToNot(HaveOccurred())

				zeroAddress := hex.EncodeToString(fabproxy.ZeroAddress)
				Expect(mockChClient.ExecuteCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.ExecuteArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         zeroAddress,
					Args:        [][]byte{[]byte(sampleArgs.Data)},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
			})
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.ExecuteReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.SendTransaction(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to execute transaction"))

				Expect(reply).To(BeEmpty())
			})
		})

		Context("when the address has a `0x` prefix", func() {
			BeforeEach(func() {
				sampleArgs.To = "0x" + sampleArgs.To
			})

			It("strips the prefix before calling the evmscc", func() {
				var reply string
				err := ethservice.SendTransaction(&http.Request{}, sampleArgs, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.ExecuteCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.ExecuteArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         sampleArgs.To[2:],
					Args:        [][]byte{[]byte(sampleArgs.Data)},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
			})
		})

		Context("when the data has a `0x` prefix", func() {
			BeforeEach(func() {
				sampleArgs.Data = "0x" + sampleArgs.Data
			})

			It("strips the prefix before calling the evmscc", func() {
				var reply string
				err := ethservice.SendTransaction(&http.Request{}, sampleArgs, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.ExecuteCallCount()).To(Equal(1))
				chReq, reqOpts := mockChClient.ExecuteArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: evmcc,
					Fcn:         sampleArgs.To,
					Args:        [][]byte{[]byte(sampleArgs.Data[2:])},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
			})
		})
	})

	/*Describe("GetTransactionReceipt", func() {
		var (
			sampleResponse      channel.Response
			sampleTransaction   *peer.ProcessedTransaction
			sampleBlock         *common.Block
			sampleTransactionID string
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{}

			var err error
			sampleTransaction, err = GetSampleTransaction([][]byte{[]byte("82373458"), []byte("sample arg 2")}, []byte("sample-response"))
			Expect(err).ToNot(HaveOccurred())

			sampleBlock, err = GetSampleBlock(1, []byte("12345abcd"))
			Expect(err).ToNot(HaveOccurred())

			mockLedgerClient.QueryBlockByTxIDReturns(sampleBlock, nil)
			mockLedgerClient.QueryTransactionReturns(sampleTransaction, nil)
			sampleTransactionID = "1234567123"
		})

		It("returns the transaction receipt associated to that transaction address", func() {
			var reply fabproxy.TxReceipt

			err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
			txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
			Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
			Expect(reqOpts).To(HaveLen(0))

			Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
			txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
			Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(fabproxy.TxReceipt{
				TransactionHash:   sampleTransactionID,
				BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
				BlockNumber:       "1",
				GasUsed:           0,
				CumulativeGasUsed: 0,
				Status:            string(uint64(1)),
			}))
		})

		Context("when the transaction is creation of a smart contract", func() {
			var contractAddress []byte
			BeforeEach(func() {
				contractAddress = []byte("0x123456789abcdef1234")
				zeroAddress := make([]byte, hex.EncodedLen(len(fabproxy.ZeroAddress)))
				hex.Encode(zeroAddress, fabproxy.ZeroAddress)

				tx, err := GetSampleTransaction([][]byte{zeroAddress, []byte("sample arg 2")}, contractAddress)
				*sampleTransaction = *tx
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the contract address in the transaction receipt", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
				txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
				Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
				Expect(reqOpts).To(HaveLen(0))

				Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
				txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
				Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(fabproxy.TxReceipt{
					TransactionHash:   sampleTransactionID,
					BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
					BlockNumber:       "1",
					ContractAddress:   string(contractAddress),
					GasUsed:           0,
					CumulativeGasUsed: 0,
					Status:            string(uint64(1)),
				}))
			})

			Context("when transaction ID has `0x` prefix", func() {
				BeforeEach(func() {
					sampleTransactionID = "0x" + sampleTransactionID
				})
				It("strips the prefix before querying the ledger", func() {
					var reply fabproxy.TxReceipt

					err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
					Expect(err).ToNot(HaveOccurred())

					Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
					txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
					Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID[2:])))
					Expect(reqOpts).To(HaveLen(0))

					Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
					txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
					Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID[2:])))
					Expect(reqOpts).To(HaveLen(0))

					Expect(reply).To(Equal(fabproxy.TxReceipt{
						TransactionHash:   sampleTransactionID,
						BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
						BlockNumber:       "1",
						ContractAddress:   string(contractAddress),
						GasUsed:           0,
						CumulativeGasUsed: 0,
						Status:            string(uint64(1)),
					}))
				})
			})
		})

		Context("when querying the ledger for the transaction errors", func() {
			BeforeEach(func() {
				mockLedgerClient.QueryTransactionReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeZero())
			})
		})

		Context("when querying the ledger for the block errors", func() {
			BeforeEach(func() {
				mockLedgerClient.QueryBlockByTxIDReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeZero())
			})
		})
	})*/

	Describe("MyNewGetTransactionReceipt", func() {
		var (
			sampleResponse      channel.Response
			sampleTransaction   *peer.ProcessedTransaction
			sampleBlock         *common.Block
			sampleTransactionID string
			msg                 events.EventDataLog
			messagePayloads     evm_event.MessagePayloads
			eventPayload        []byte
			eventBytes          []byte
			sampleAddress       string
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{}

			var err error

			sampleTransactionID = "1234567123"

			sampleAddress = "0000000000000address"
			addr, er := account.AddressFromBytes([]byte(sampleAddress))
			Expect(er).ToNot(HaveOccurred())

			msg = events.EventDataLog{
				Address: addr,
				Topics:  []binary.Word256{[32]byte{0x7, 0x79, 0x9c, 0x56, 0x12, 0x2d, 0x95, 0x24, 0x5a, 0xc7, 0x9c, 0xa1, 0x71, 0xa8, 0xd0, 0x25, 0xdc, 0x20, 0x33, 0x2c, 0xcf, 0xf9, 0x54, 0x8, 0xde, 0x17, 0xbc, 0xaa, 0x73, 0xc8, 0xca, 0x1c}, [32]byte{0xec, 0xa6, 0x62, 0xca, 0xe7, 0x47, 0xb4, 0x67, 0x82, 0x2a, 0x1d, 0x79, 0xb1, 0xeb, 0x1a, 0xee, 0xf1, 0x3b, 0xff, 0x8c, 0x77, 0x39, 0x44, 0x34, 0x46, 0xd4, 0xfd, 0x74, 0xfb, 0x15, 0x12, 0x5f}},
				Data:    []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x20, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x10},
				Height:  0,
			}
			messagePayloads.Payloads = make([]evm_event.MessagePayload, 0)
			messagePayloads.Payloads = append(messagePayloads.Payloads, evm_event.MessagePayload{Message: msg})
			eventPayload, err = json.Marshal(messagePayloads)
			Expect(err).ToNot(HaveOccurred())

			chaincodeEvent := peer.ChaincodeEvent{
				ChaincodeId: "qscc",
				TxId:        sampleTransactionID,
				EventName:   "Chaincode event",
				Payload:     eventPayload,
			}

			eventBytes, err = proto.Marshal(&chaincodeEvent)
			Expect(err).ToNot(HaveOccurred())

			sampleTransaction, err = GetSampleTransaction([][]byte{[]byte("82373458"), []byte("sample arg 2")}, eventBytes, []byte("sample-response"))
			Expect(err).ToNot(HaveOccurred())

			sampleBlock, err = GetSampleBlock(1, []byte("12345abcd"))
			Expect(err).ToNot(HaveOccurred())

			mockLedgerClient.QueryBlockByTxIDReturns(sampleBlock, nil)
			mockLedgerClient.QueryTransactionReturns(sampleTransaction, nil)
		})

		It("returns the transaction receipt associated to that transaction address", func() {
			var reply fabproxy.TxReceipt

			err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
			txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
			Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
			Expect(reqOpts).To(HaveLen(0))

			Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
			txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
			Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
			Expect(reqOpts).To(HaveLen(0))

			var topics []string
			topics = make([]string, 0)
			for _, topic := range msg.Topics {
				topics = append(topics, topic.String())
			}

			expectedLog := fabproxy.Log{
				Address:     hex.EncodeToString([]byte(sampleAddress)),
				Topics:      topics,
				Data:        string(msg.Data),
				BlockNumber: "1",
				TxHash:      sampleTransactionID,
				//TxIndex: ,
				BlockHash: hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
				Index:     string(0),
				Type:      "mined",
			}

			var expectedLogs []fabproxy.Log
			expectedLogs = make([]fabproxy.Log, 0)
			expectedLogs = append(expectedLogs, expectedLog)

			var expectedBloom fabproxy.Bloom
			expectedBloom = fabproxy.CreateBloom(expectedLogs)

			Expect(reply).To(Equal(fabproxy.TxReceipt{
				TransactionHash:   sampleTransactionID,
				BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
				BlockNumber:       "1",
				GasUsed:           0,
				CumulativeGasUsed: 0,
				Logs:              expectedLogs,
				LogsBloom:         expectedBloom,
				Status:            string(uint64(1)),
			}))
		})

		Context("when the transaction is creation of a smart contract", func() {
			var contractAddress []byte
			BeforeEach(func() {
				contractAddress = []byte("0x123456789abcdef1234")
				zeroAddress := make([]byte, hex.EncodedLen(len(fabproxy.ZeroAddress)))
				hex.Encode(zeroAddress, fabproxy.ZeroAddress)
				tx, err := GetSampleTransaction([][]byte{zeroAddress, []byte("sample arg 2")}, nil, contractAddress)
				*sampleTransaction = *tx
				Expect(err).ToNot(HaveOccurred())

			})

			It("returns the contract address in the transaction receipt", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
				txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
				Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
				Expect(reqOpts).To(HaveLen(0))

				Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
				txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
				Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID)))
				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(fabproxy.TxReceipt{
					TransactionHash:   sampleTransactionID,
					BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
					BlockNumber:       "1",
					ContractAddress:   string(contractAddress),
					GasUsed:           0,
					CumulativeGasUsed: 0,
					Logs:              nil,
					LogsBloom:         fabproxy.CreateBloom(nil),
					Status:            string(uint64(1)),
				}))
			})

			Context("when transaction ID has `0x` prefix", func() {
				BeforeEach(func() {
					sampleTransactionID = "0x" + sampleTransactionID
				})
				It("strips the prefix before querying the ledger", func() {
					var reply fabproxy.TxReceipt

					err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
					Expect(err).ToNot(HaveOccurred())

					Expect(mockLedgerClient.QueryTransactionCallCount()).To(Equal(1))
					txID, reqOpts := mockLedgerClient.QueryTransactionArgsForCall(0)
					Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID[2:])))
					Expect(reqOpts).To(HaveLen(0))

					Expect(mockLedgerClient.QueryBlockByTxIDCallCount()).To(Equal(1))
					txID, reqOpts = mockLedgerClient.QueryBlockByTxIDArgsForCall(0)
					Expect(txID).To(Equal(fab.TransactionID(sampleTransactionID[2:])))
					Expect(reqOpts).To(HaveLen(0))

					Expect(reply).To(Equal(fabproxy.TxReceipt{
						TransactionHash:   sampleTransactionID,
						BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().GetDataHash()),
						BlockNumber:       "1",
						ContractAddress:   string(contractAddress),
						GasUsed:           0,
						CumulativeGasUsed: 0,
						Logs:              nil,
						LogsBloom:         fabproxy.CreateBloom(nil),
						Status:            string(uint64(1)),
					}))
				})
			})

			Context("when querying the ledger for the transaction errors", func() {
				BeforeEach(func() {
					mockLedgerClient.QueryTransactionReturns(nil, errors.New("boom!"))
				})

				It("returns a corresponding error", func() {
					var reply fabproxy.TxReceipt

					err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

					Expect(reply).To(BeZero())
				})
			})

			Context("when querying the ledger for the block errors", func() {
				BeforeEach(func() {
					mockLedgerClient.QueryBlockByTxIDReturns(nil, errors.New("boom!"))
				})

				It("returns a corresponding error", func() {
					var reply fabproxy.TxReceipt

					err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

					Expect(reply).To(BeZero())
				})
			})
		})
	})

	Describe("Accounts", func() {
		var (
			sampleAccount string
			arg           string
		)

		BeforeEach(func() {
			sampleAccount = "123456ABCD"
			mockChClient.QueryReturns(channel.Response{
				Payload: []byte(sampleAccount),
			}, nil)

		})

		It("requests the user address from the evmscc based on the user cert", func() {
			var reply []string

			err := ethservice.Accounts(&http.Request{}, &arg, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(1))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: evmcc,
				Fcn:         "account",
				Args:        [][]byte{},
			}))

			Expect(reqOpts).To(HaveLen(0))

			expectedResponse := []string{"0x" + strings.ToLower(sampleAccount)}

			Expect(reply).To(Equal(expectedResponse))
		})

		Context("when querying the ledger errors", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply []string

				err := ethservice.Accounts(&http.Request{}, &arg, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeEmpty())
			})
		})
	})
})

func GetSampleBlock(blkNumber uint64, blkHash []byte) (*common.Block, error) {
	return &common.Block{
		Header: &common.BlockHeader{Number: blkNumber, DataHash: blkHash},
	}, nil
}

func GetSampleTransaction(inputArgs [][]byte, eventBytes []byte, txResponse []byte) (*peer.ProcessedTransaction, error) {

	respPayload := &peer.ChaincodeAction{
		Events: eventBytes,
		Response: &peer.Response{
			Payload: txResponse,
		},
	}

	ext, err := proto.Marshal(respPayload)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	pRespPayload := &peer.ProposalResponsePayload{
		Extension: ext,
	}

	ccProposalPayload, err := proto.Marshal(pRespPayload)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{
				Name: evmcc,
			},
			Input: &peer.ChaincodeInput{
				Args: inputArgs,
			},
		},
	}

	invokeSpecBytes, err := proto.Marshal(invokeSpec)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	ccPropPayload, err := proto.Marshal(&peer.ChaincodeProposalPayload{
		Input: invokeSpecBytes,
	})
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	ccPayload := &peer.ChaincodeActionPayload{
		Action: &peer.ChaincodeEndorsedAction{
			ProposalResponsePayload: ccProposalPayload,
		},
		ChaincodeProposalPayload: ccPropPayload,
	}

	actionPayload, err := proto.Marshal(ccPayload)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	txAction := &peer.TransactionAction{
		Payload: actionPayload,
	}

	txActions := &peer.Transaction{
		Actions: []*peer.TransactionAction{txAction},
	}

	actionsPayload, err := proto.Marshal(txActions)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	payload := &common.Payload{
		Data: actionsPayload,
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	tx := &peer.ProcessedTransaction{
		TransactionEnvelope: &common.Envelope{
			Payload: payloadBytes,
		},
	}

	return tx, nil
}
