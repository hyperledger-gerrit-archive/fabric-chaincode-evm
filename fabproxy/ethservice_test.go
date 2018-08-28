/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy_test

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
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

		Context("when the ledger errors when processing a query", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to query the ledger")))

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

		Context("when the ledger errors when processing a query", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to query the ledger")))
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

		Context("when the ledger errors when processing a query", func() {
			BeforeEach(func() {
				mockChClient.ExecuteReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.SendTransaction(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to execute transaction")))
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

	Describe("GetTransactionReceipt", func() {
		var (
			sampleTransaction   *peer.ProcessedTransaction
			sampleBlock         *common.Block
			sampleTransactionID string
		)

		BeforeEach(func() {
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
					}))
				})
			})
		})

		Context("when the ledger errors when processing a transaction query for the transaction", func() {
			BeforeEach(func() {
				mockLedgerClient.QueryTransactionReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to query the ledger")))
				Expect(reply).To(BeZero())
			})
		})

		Context("when the ledger errors when processing a query for the block", func() {
			BeforeEach(func() {
				mockLedgerClient.QueryBlockByTxIDReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to query the ledger")))
				Expect(reply).To(BeZero())
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

		Context("when the ledger errors when processing a query", func() {
			BeforeEach(func() {
				mockChClient.QueryReturns(channel.Response{}, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply []string
				err := ethservice.Accounts(&http.Request{}, &arg, &reply)
				Expect(err).To(MatchError(ContainSubstring("Failed to query the ledger")))
				Expect(reply).To(BeEmpty())
			})
		})
	})

	Describe("EstimateGas", func() {
		It("always returns zero", func() {
			var reply string
			err := ethservice.EstimateGas(&http.Request{}, &fabproxy.EthArgs{}, &reply)
			Expect(err).ToNot(HaveOccurred())
			Expect(reply).To(Equal("0x0"))
		})
	})

	Describe("GetBalance", func() {
		It("always returns zero", func() {
			arg := make([]string, 2)
			var reply string
			err := ethservice.GetBalance(&http.Request{}, &arg, &reply)
			Expect(err).ToNot(HaveOccurred())
			Expect(reply).To(Equal("0x0"))
		})
	})

	Describe("GetBlockByNumber", func() {

		gst := func(inputArgs [][]byte, txResponse []byte) *peer.ProcessedTransaction {
			t, e := GetSampleTransaction(inputArgs, txResponse)
			Expect(e).ToNot(HaveOccurred())
			return t
		}

		Context("bad parameters", func() {
			Context("an incorrect number of args", func() {
				var reply fabproxy.Block
				Specify("no args", func() {
					var arg []interface{}
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				Specify("one arg", func() {
					arg := make([]interface{}, 1)
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				Specify("more than two args", func() {
					arg := make([]interface{}, 3)
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
			})
			Context("wrong arg types", func() {
				var reply fabproxy.Block
				Specify("not a string as first arg", func() {
					arg := make([]interface{}, 2)
					arg[0] = false
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				Specify("not a named block or numbered block as first arg", func() {
					arg := make([]interface{}, 2)
					arg[0] = "hurf%&"
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				Specify("not a boolean as second arg", func() {
					arg := make([]interface{}, 2)
					arg[0] = "latest"
					arg[1] = "durf"
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("good parameters", func() {
			var reply fabproxy.Block
			Context("partial transactions", func() {
				It("ledger client fails block by number", func() {
					arg := make([]interface{}, 2)
					arg[0] = "0x0"
					arg[1] = false

					mockLedgerClient.QueryBlockReturns(nil, fmt.Errorf("no block"))
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				It("ledger client fails block by name with no blockchain info", func() {
					arg := make([]interface{}, 2)
					arg[0] = "latest"
					arg[1] = false

					mockLedgerClient.QueryInfoReturns(nil, fmt.Errorf("no block info"))
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				It("ledger client fails block by name latest no block by number", func() {
					arg := make([]interface{}, 2)
					arg[0] = "latest"
					arg[1] = false
					mockLedgerClient.QueryInfoReturns(&fab.BlockchainInfoResponse{BCI: &common.BlockchainInfo{Height: 1}}, nil)
					mockLedgerClient.QueryBlockReturns(nil, fmt.Errorf("no block"))
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				It("ledger client fails block by name eaest no block by number", func() {
					arg := make([]interface{}, 2)
					arg[0] = "earliest"
					arg[1] = false
					mockLedgerClient.QueryInfoReturns(&fab.BlockchainInfoResponse{BCI: &common.BlockchainInfo{Height: 1}}, nil)
					mockLedgerClient.QueryBlockReturns(nil, fmt.Errorf("no block"))
					err := ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).To(HaveOccurred())
				})
				It("requests a block by number", func() {
					blocknumber := "abc0"
					arg := make([]interface{}, 2)
					arg[0] = blocknumber
					arg[1] = false

					txn1, _ := proto.Marshal(gst([][]byte{[]byte("12345678"), []byte("sample arg 1")}, []byte("sample-response1")).TransactionEnvelope)
					txn2, _ := proto.Marshal(gst([][]byte{[]byte("98765432"), []byte("sample arg 2")}, []byte("sample-response2")).TransactionEnvelope)

					ublocknumber, err := strconv.ParseUint(blocknumber, 16, 64)
					phash := []byte("abc\x00")
					dhash := []byte("def\xFF")
					mockLedgerClient.QueryBlockReturns(&common.Block{
						Header: &common.BlockHeader{Number: ublocknumber,
							PreviousHash: phash,
							DataHash:     dhash},
						Data: &common.BlockData{Data: [][]byte{txn1, txn2}},
					},
						nil)

					err = ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).ToNot(HaveOccurred())
					Expect(reply.Number).To(Equal("0x"+blocknumber), "block number")
					Expect(reply.Hash).To(Equal("0x"+hex.EncodeToString(dhash)), "block data hash")
					Expect(reply.ParentHash).To(Equal("0x"+hex.EncodeToString(phash)), "block parent hash")
					txns := reply.Transactions
					Expect(txns).To(HaveLen(2))
					Expect(txns[0]).To(BeEquivalentTo("0x1234"))
					Expect(txns[1]).To(BeEquivalentTo("0x1234"))

				})
				It("requests a block by name", func() {
					blocknumber := "abc0"
					arg := make([]interface{}, 2)
					arg[0] = "latest"
					arg[1] = false
					ublocknumber, err := strconv.ParseUint(blocknumber, 16, 64)
					By(fmt.Sprint(ublocknumber))
					mockLedgerClient.QueryInfoReturns(&fab.BlockchainInfoResponse{BCI: &common.BlockchainInfo{Height: ublocknumber + 1}}, nil)

					txn1, _ := proto.Marshal(gst([][]byte{[]byte("12345678"), []byte("sample arg 1")}, []byte("sample-response1")).TransactionEnvelope)
					txn2, _ := proto.Marshal(gst([][]byte{[]byte("98765432"), []byte("sample arg 2")}, []byte("sample-response2")).TransactionEnvelope)

					phash := []byte("abc\x00")
					dhash := []byte("def\xFF")
					mockLedgerClient.QueryBlockReturns(&common.Block{
						Header: &common.BlockHeader{Number: ublocknumber,
							PreviousHash: phash,
							DataHash:     dhash},
						Data: &common.BlockData{Data: [][]byte{txn1, txn2}},
					},
						nil)

					err = ethservice.GetBlockByNumber(&http.Request{}, &arg, &reply)
					Expect(err).ToNot(HaveOccurred())
					Expect(reply.Number).To(Equal("0x"+blocknumber), "block number")
					Expect(reply.Hash).To(Equal("0x"+hex.EncodeToString(dhash)), "block data hash")
					Expect(reply.ParentHash).To(Equal("0x"+hex.EncodeToString(phash)), "block parent hash")
					txns := reply.Transactions
					Expect(txns).To(HaveLen(2))
					Expect(txns[0]).To(BeEquivalentTo("0x1234"))
					Expect(txns[1]).To(BeEquivalentTo("0x1234"))

				})
			})
		})
	})
})

func GetSampleBlock(blkNumber uint64, blkHash []byte) (*common.Block, error) {
	return &common.Block{
		Header: &common.BlockHeader{Number: blkNumber, DataHash: blkHash},
	}, nil
}

func GetSampleTransaction(inputArgs [][]byte, txResponse []byte) (*peer.ProcessedTransaction, error) {

	respPayload := &peer.ChaincodeAction{
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

	chdr := &common.ChannelHeader{TxId: "1234"}
	chdrBytes, err := proto.Marshal(chdr)
	if err != nil {
		return &peer.ProcessedTransaction{}, err
	}

	payload := &common.Payload{
		Header: &common.Header{
			ChannelHeader: chdrBytes,
		},
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
