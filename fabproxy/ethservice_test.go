package fabproxy_test

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	"github.com/hyperledger/fabric-chaincode-evm/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var channelID = "mychannel"
var _ = Describe("Ethservice", func() {
	var (
		ethservice fabproxy.EthService

		fabSDK       *mocks.MockSDK
		mockChClient *mocks.MockChannelClient
		channelID    string
	)

	BeforeEach(func() {
		fabSDK = &mocks.MockSDK{}
		mockChClient = &mocks.MockChannelClient{}
		channelID = "test-channel"

		fabSDK.GetChannelClientReturns(mockChClient, nil)
		ethservice = fabproxy.NewEthService(fabSDK, channelID)
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
				ChaincodeID: fabproxy.EVMSCC,
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
					ChaincodeID: fabproxy.EVMSCC,
					Fcn:         "getCode",
					Args:        [][]byte{[]byte(sampleAddress[2:])},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleCode)))
			})
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.GetCode(&http.Request{}, &sampleAddress, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
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
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         sampleArgs.To,
				Args:        [][]byte{[]byte(sampleArgs.Data)},
			}))

			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal("0x" + string(encodedResponse)))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.Call(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
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
					ChaincodeID: fabproxy.EVMSCC,
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
					ChaincodeID: fabproxy.EVMSCC,
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
				ChaincodeID: fabproxy.EVMSCC,
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
					ChaincodeID: fabproxy.EVMSCC,
					Fcn:         zeroAddress,
					Args:        [][]byte{[]byte(sampleArgs.Data)},
				}))

				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(string(sampleResponse.TransactionID)))
			})
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply string

				err := ethservice.SendTransaction(&http.Request{}, &fabproxy.EthArgs{}, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
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
					ChaincodeID: fabproxy.EVMSCC,
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
					ChaincodeID: fabproxy.EVMSCC,
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
			sampleResponse      channel.Response
			sampleTransaction   peer.ProcessedTransaction
			sampleBlock         common.Block
			sampleTransactionID string
			txBytes             []byte
			blkBytes            []byte
		)

		BeforeEach(func() {
			sampleResponse = channel.Response{}

			var err error
			sampleTransaction, err = GetSampleTransaction([][]byte{[]byte("82373458"), []byte("sample arg 2")}, []byte("sample-response"))
			Expect(err).ToNot(HaveOccurred())

			sampleBlock, err = GetSampleBlock(1, []byte("12345abcd"))
			Expect(err).ToNot(HaveOccurred())

			mockChClient.QueryStub = func(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {

				if request.Fcn == "GetTransactionByID" {
					sampleResponse.Payload = txBytes

				} else if request.Fcn == "GetBlockByTxID" {
					sampleResponse.Payload = blkBytes
				}

				return sampleResponse, nil
			}

			sampleTransactionID = "1234567123"
		})

		JustBeforeEach(func() {
			var err error
			txBytes, err = proto.Marshal(&sampleTransaction)
			Expect(err).ToNot(HaveOccurred())

			blkBytes, err = proto.Marshal(&sampleBlock)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns the transaction receipt associated to that transaction address", func() {
			var reply fabproxy.TxReceipt

			err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
			Expect(err).ToNot(HaveOccurred())

			Expect(mockChClient.QueryCallCount()).To(Equal(2))
			chReq, reqOpts := mockChClient.QueryArgsForCall(0)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.QSCC,
				Fcn:         "GetTransactionByID",
				Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID)},
			}))
			Expect(reqOpts).To(HaveLen(0))

			chReq, reqOpts = mockChClient.QueryArgsForCall(1)
			Expect(chReq).To(Equal(channel.Request{
				ChaincodeID: fabproxy.QSCC,
				Fcn:         "GetBlockByTxID",
				Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID)},
			}))
			Expect(reqOpts).To(HaveLen(0))

			Expect(reply).To(Equal(fabproxy.TxReceipt{
				TransactionHash:   sampleTransactionID,
				BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().Hash()),
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
				var err error
				sampleTransaction, err = GetSampleTransaction([][]byte{zeroAddress, []byte("sample arg 2")}, contractAddress)
				Expect(err).ToNot(HaveOccurred())

			})

			It("returns the contract address in the transaction receipt", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).ToNot(HaveOccurred())

				Expect(mockChClient.QueryCallCount()).To(Equal(2))
				chReq, reqOpts := mockChClient.QueryArgsForCall(0)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: fabproxy.QSCC,
					Fcn:         "GetTransactionByID",
					Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID)},
				}))
				Expect(reqOpts).To(HaveLen(0))

				chReq, reqOpts = mockChClient.QueryArgsForCall(1)
				Expect(chReq).To(Equal(channel.Request{
					ChaincodeID: fabproxy.QSCC,
					Fcn:         "GetBlockByTxID",
					Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID)},
				}))
				Expect(reqOpts).To(HaveLen(0))

				Expect(reply).To(Equal(fabproxy.TxReceipt{
					TransactionHash:   sampleTransactionID,
					BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().Hash()),
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

					Expect(mockChClient.QueryCallCount()).To(Equal(2))
					chReq, reqOpts := mockChClient.QueryArgsForCall(0)
					Expect(chReq).To(Equal(channel.Request{
						ChaincodeID: fabproxy.QSCC,
						Fcn:         "GetTransactionByID",
						Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID[2:])},
					}))
					Expect(reqOpts).To(HaveLen(0))

					chReq, reqOpts = mockChClient.QueryArgsForCall(1)
					Expect(chReq).To(Equal(channel.Request{
						ChaincodeID: fabproxy.QSCC,
						Fcn:         "GetBlockByTxID",
						Args:        [][]byte{[]byte(channelID), []byte(sampleTransactionID[2:])},
					}))
					Expect(reqOpts).To(HaveLen(0))

					Expect(reply).To(Equal(fabproxy.TxReceipt{
						TransactionHash:   sampleTransactionID,
						BlockHash:         hex.EncodeToString(sampleBlock.GetHeader().Hash()),
						BlockNumber:       "1",
						ContractAddress:   string(contractAddress),
						GasUsed:           0,
						CumulativeGasUsed: 0,
					}))
				})
			})
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt
				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeZero())
			})
		})

		// Need to test multiple places the querying can fail
		Context("when querying the ledger for the transaction errors", func() {
			BeforeEach(func() {
				mockChClient.QueryStub = func(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {

					if request.Fcn == "GetTransactionByID" {
						return sampleResponse, errors.New("boom!")
					} else if request.Fcn == "GetBlockByTxID" {

						txBytes, err := proto.Marshal(&sampleBlock)
						Expect(err).ToNot(HaveOccurred())
						sampleResponse.Payload = txBytes
					}

					return sampleResponse, nil
				}
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
				mockChClient.QueryStub = func(request channel.Request, options ...channel.RequestOption) (channel.Response, error) {

					if request.Fcn == "GetTransactionByID" {

						txBytes, err := proto.Marshal(&sampleTransaction)
						Expect(err).ToNot(HaveOccurred())
						sampleResponse.Payload = txBytes

					} else if request.Fcn == "GetBlockByTxID" {
						return sampleResponse, errors.New("boom!")
					}

					return sampleResponse, nil
				}
			})

			It("returns a corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to query the ledger"))

				Expect(reply).To(BeZero())
			})
		})

		Context("when decoding the args of the transaction fails", func() {
			BeforeEach(func() {
				var err error
				sampleTransaction, err = GetSampleTransaction([][]byte{[]byte("sample arg1"), []byte("sample arg 2")}, []byte("sample-response"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the corresponding error", func() {

				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to decode transaction arguments"))

				Expect(reply).To(BeZero())
			})
		})

		Context("when the transaction payload is malformed", func() {
			JustBeforeEach(func() {
				txBytes = append(txBytes, '0')
			})

			It("returns the corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to unmarshal transaction"))

				Expect(reply).To(BeZero())
			})
		})

		Context("when the transaction payload is malformed", func() {
			JustBeforeEach(func() {
				blkBytes = append(blkBytes, '0')
			})

			It("returns the corresponding error", func() {
				var reply fabproxy.TxReceipt

				err := ethservice.GetTransactionReceipt(&http.Request{}, &sampleTransactionID, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to unmarshal block"))

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
				ChaincodeID: fabproxy.EVMSCC,
				Fcn:         "account",
				Args:        [][]byte{},
			}))

			Expect(reqOpts).To(HaveLen(0))

			expectedResponse := []string{"0x" + strings.ToLower(sampleAccount)}

			Expect(reply).To(Equal(expectedResponse))
		})

		Context("when getting the channel client errors ", func() {
			BeforeEach(func() {
				fabSDK.GetChannelClientReturns(nil, errors.New("boom!"))
			})

			It("returns a corresponding error", func() {
				var reply []string

				err := ethservice.Accounts(&http.Request{}, &arg, &reply)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Failed to generate channel client"))

				Expect(reply).To(BeEmpty())
			})
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

func GetSampleBlock(blkNumber uint64, blkHash []byte) (common.Block, error) {
	blk := common.Block{
		Header: &common.BlockHeader{Number: blkNumber, DataHash: blkHash},
	}

	return blk, nil
}

func GetSampleTransaction(inputArgs [][]byte, txResponse []byte) (peer.ProcessedTransaction, error) {

	respPayload := &peer.ChaincodeAction{
		Response: &peer.Response{
			Payload: txResponse,
		},
	}

	ext, err := proto.Marshal(respPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	pRespPayload := &peer.ProposalResponsePayload{
		Extension: ext,
	}

	ccProposalPayload, err := proto.Marshal(pRespPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{
				Name: fabproxy.EVMSCC,
			},
			Input: &peer.ChaincodeInput{
				Args: inputArgs,
			},
		},
	}

	invokeSpecBytes, err := proto.Marshal(invokeSpec)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	ccPropPayload, err := proto.Marshal(&peer.ChaincodeProposalPayload{
		Input: invokeSpecBytes,
	})
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	ccPayload := &peer.ChaincodeActionPayload{
		Action: &peer.ChaincodeEndorsedAction{
			ProposalResponsePayload: ccProposalPayload,
		},
		ChaincodeProposalPayload: ccPropPayload,
	}

	actionPayload, err := proto.Marshal(ccPayload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	txAction := &peer.TransactionAction{
		Payload: actionPayload,
	}

	txActions := &peer.Transaction{
		Actions: []*peer.TransactionAction{txAction},
	}

	actionsPayload, err := proto.Marshal(txActions)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	payload := &common.Payload{
		Data: actionsPayload,
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return peer.ProcessedTransaction{}, err
	}

	tx := peer.ProcessedTransaction{
		TransactionEnvelope: &common.Envelope{
			Payload: payloadBytes,
		},
	}

	return tx, nil
}
