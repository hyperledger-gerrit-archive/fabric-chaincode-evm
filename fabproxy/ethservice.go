/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var ZeroAddress = make([]byte, 20)

//go:generate counterfeiter -o ../mocks/mockchannelclient.go --fake-name MockChannelClient ./ ChannelClient
type ChannelClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

//go:generate counterfeiter -o ../mocks/mockledgerclient.go --fake-name MockLedgerClient ./ LedgerClient
type LedgerClient interface {
	QueryInfo(options ...ledger.RequestOption) (*fab.BlockchainInfoResponse, error)
	QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*common.Block, error)
	QueryBlockByTxID(txid fab.TransactionID, options ...ledger.RequestOption) (*common.Block, error)
	QueryTransaction(txid fab.TransactionID, options ...ledger.RequestOption) (*peer.ProcessedTransaction, error)
}

// EthService is the rpc server implementation. Each function is an
// implementation of one ethereum json-rpc
// https://github.com/ethereum/wiki/wiki/JSON-RPC
//
// Arguments and return values are formatted as HEX value encoding
// https://github.com/ethereum/wiki/wiki/JSON-RPC#hex-value-encoding
//
//go:generate counterfeiter -o ../mocks/mockethservice.go --fake-name MockEthService ./ EthService
type EthService interface {
	GetCode(r *http.Request, arg *string, reply *string) error
	Call(r *http.Request, args *EthArgs, reply *string) error
	SendTransaction(r *http.Request, args *EthArgs, reply *string) error
	GetTransactionReceipt(r *http.Request, arg *string, reply *TxReceipt) error
	Accounts(r *http.Request, arg *string, reply *[]string) error
	EstimateGas(r *http.Request, args *EthArgs, reply *string) error
	GetBalance(r *http.Request, p *[]string, reply *string) error
	GetBlockByNumber(r *http.Request, p *[]interface{}, reply *Block) error
	GetTransactionByHash(r *http.Request, txID *string, reply *Transaction) error
}

type ethService struct {
	channelClient ChannelClient
	ledgerClient  LedgerClient
	channelID     string
	ccid          string
}

type EthArgs struct {
	To       string `json:"to"`
	From     string `json:"from"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	Nonce    string `json:"nonce"`
}

type TxReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	ContractAddress   string `json:"contractAddress"`
	GasUsed           int    `json:"gasUsed"`
	CumulativeGasUsed int    `json:"cumulativeGasUsed"`
}

func NewEthService(channelClient ChannelClient, ledgerClient LedgerClient, channelID string, ccid string) EthService {
	return &ethService{channelClient: channelClient, ledgerClient: ledgerClient, channelID: channelID, ccid: ccid}
}

func (s *ethService) GetCode(r *http.Request, arg *string, reply *string) error {
	strippedAddr := strip0x(*arg)

	response, err := s.query(s.ccid, "getCode", [][]byte{[]byte(strippedAddr)})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	*reply = string(response.Payload)

	return nil
}

func (s *ethService) Call(r *http.Request, args *EthArgs, reply *string) error {
	response, err := s.query(s.ccid, strip0x(args.To), [][]byte{[]byte(strip0x(args.Data))})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	// Clients expect the prefix to present in responses
	*reply = "0x" + hex.EncodeToString(response.Payload)

	return nil
}

func (s *ethService) SendTransaction(r *http.Request, args *EthArgs, reply *string) error {
	if args.To == "" {
		args.To = hex.EncodeToString(ZeroAddress)
	}

	response, err := s.channelClient.Execute(channel.Request{
		ChaincodeID: s.ccid,
		Fcn:         strip0x(args.To),
		Args:        [][]byte{[]byte(strip0x(args.Data))},
	})

	if err != nil {
		return errors.New(fmt.Sprintf("Failed to execute transaction: %s", err.Error()))
	}

	*reply = string(response.TransactionID)
	return nil
}

func (s *ethService) GetTransactionReceipt(r *http.Request, txID *string, reply *TxReceipt) error {
	strippedTxId := strip0x(*txID)

	args := [][]byte{[]byte(s.channelID), []byte(strippedTxId)}

	tx, err := s.ledgerClient.QueryTransaction(fab.TransactionID(strippedTxId))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	p := tx.GetTransactionEnvelope().GetPayload()
	payload := &common.Payload{}
	err = proto.Unmarshal(p, payload)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	txActions := &peer.Transaction{}
	err = proto.Unmarshal(payload.GetData(), txActions)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	ccPropPayload, respPayload, err := getPayloads(txActions.GetActions()[0])
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(ccPropPayload.GetInput(), invokeSpec)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to unmarshal transaction: %s", err.Error()))
	}

	block, err := s.ledgerClient.QueryBlockByTxID(fab.TransactionID(strippedTxId))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	blkHeader := block.GetHeader()

	receipt := TxReceipt{
		TransactionHash:   *txID,
		BlockHash:         hex.EncodeToString(blkHeader.GetDataHash()),
		BlockNumber:       strconv.FormatUint(blkHeader.GetNumber(), 10),
		GasUsed:           0,
		CumulativeGasUsed: 0,
	}

	args = invokeSpec.GetChaincodeSpec().GetInput().Args
	// First arg is the callee address. If it is zero address, tx was a contract creation
	callee, err := hex.DecodeString(string(args[0]))
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to decode transaction arguments: %s", err.Error()))
	}

	if bytes.Equal(callee, ZeroAddress) {
		receipt.ContractAddress = string(respPayload.GetResponse().GetPayload())
	}
	*reply = receipt

	return nil
}

func (s *ethService) Accounts(r *http.Request, arg *string, reply *[]string) error {
	response, err := s.query(s.ccid, "account", [][]byte{})
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to query the ledger: %s", err.Error()))
	}

	*reply = []string{"0x" + strings.ToLower(string(response.Payload))}

	return nil
}

func (s *ethService) query(ccid, function string, queryArgs [][]byte) (channel.Response, error) {

	return s.channelClient.Query(channel.Request{
		ChaincodeID: ccid,
		Fcn:         function,
		Args:        queryArgs,
	})
}

func strip0x(addr string) string {
	//Not checking for malformed addresses just stripping `0x` prefix where applicable
	if len(addr) > 2 && addr[0:2] == "0x" {
		return addr[2:]
	}
	return addr
}

func getPayloads(txActions *peer.TransactionAction) (*peer.ChaincodeProposalPayload, *peer.ChaincodeAction, error) {
	// TODO: pass in the tx type (in what follows we're assuming the type is ENDORSER_TRANSACTION)
	ccPayload := &peer.ChaincodeActionPayload{}
	err := proto.Unmarshal(txActions.Payload, ccPayload)
	if err != nil {
		return nil, nil, err
	}

	if ccPayload.Action == nil || ccPayload.Action.ProposalResponsePayload == nil {
		return nil, nil, fmt.Errorf("no payload in ChaincodeActionPayload")
	}

	ccProposalPayload := &peer.ChaincodeProposalPayload{}
	err = proto.Unmarshal(ccPayload.ChaincodeProposalPayload, ccProposalPayload)
	if err != nil {
		return nil, nil, err
	}

	pRespPayload := &peer.ProposalResponsePayload{}
	err = proto.Unmarshal(ccPayload.Action.ProposalResponsePayload, pRespPayload)
	if err != nil {
		return nil, nil, err
	}

	if pRespPayload.Extension == nil {
		return nil, nil, fmt.Errorf("response payload is missing extension")
	}

	respPayload := &peer.ChaincodeAction{}
	err = proto.Unmarshal(pRespPayload.Extension, respPayload)
	if err != nil {
		return ccProposalPayload, nil, err
	}
	return ccProposalPayload, respPayload, nil
}

// EstimateGas accepts the same arguments as Call but all arguments are
// optional.  This implementation ignores all arguments and returns a zero
// estimate.
//
// The intention is to estimate how much gas is necessary to allow a transaction
// to complete.
//
// EVM-chaincode does not require gas to run transactions. The chaincode will
// give enough gas per transaction.
func (s *ethService) EstimateGas(r *http.Request, _ *EthArgs, reply *string) error {
	fmt.Println("EstimateGas called")
	*reply = "0x0"
	return nil
}

// GetBalance takes an address and a block, but this implementation
// does not check or use either parameter.
//
// Always returns zero.
func (s *ethService) GetBalance(r *http.Request, p *[]string, reply *string) error {
	fmt.Println("GetBalance called")
	*reply = "0x0"
	return nil
}

// Block is an eth return struct
// defined https://github.com/ethereum/wiki/wiki/JSON-RPC#returns-26
type Block struct {
	Number     string `json:"number"`     // number: QUANTITY - the block number. null when its pending block.
	Hash       string `json:"hash"`       // hash: DATA, 32 Bytes - hash of the block. null when its pending block.
	ParentHash string `json:"parentHash"` // parentHash: DATA, 32 Bytes - hash of the parent block.
	// nonce: DATA, 8 Bytes - hash of the generated proof-of-work. null when its pending block.
	// sha3Uncles: DATA, 32 Bytes - SHA3 of the uncles data in the block.
	// logsBloom: DATA, 256 Bytes - the bloom filter for the logs of the block. null when its pending block.
	// transactionsRoot: DATA, 32 Bytes - the root of the transaction trie of the block.
	// stateRoot: DATA, 32 Bytes - the root of the final state trie of the block.
	// receiptsRoot: DATA, 32 Bytes - the root of the receipts trie of the block.
	// miner: DATA, 20 Bytes - the address of the beneficiary to whom the mining rewards were given.
	// difficulty: QUANTITY - integer of the difficulty for this block.
	// totalDifficulty: QUANTITY - integer of the total difficulty of the chain until this block.
	// extraData: DATA - the "extra data" field of this block.
	// size: QUANTITY - integer the size of this block in bytes.
	// gasLimit: QUANTITY - the maximum gas allowed in this block.
	// gasUsed: QUANTITY - the total used gas by all transactions in this block.
	// timestamp: QUANTITY - the unix timestamp for when the block was collated.
	Transactions []interface{} `json:"transactions"` // transactions: Array - Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter.
	//  uncles: Array - Array of uncle hashes.
}

// https://github.com/ethereum/wiki/wiki/JSON-RPC#the-default-block-parameter
func parseAsDefaultBlock(input string) (*defaultBlock, error) {
	// check if it's one of the nameed-blocks
	if input == "latest" || input == "earliest" || input == "pending" {
		return &defaultBlock{namedBlock: input}, nil
	}
	// check if it's a number
	// RPC defines it as a hex-string (could use 0 middle arg for more lenient parsing)
	decoded, parseErr := strconv.ParseUint(input, 16, 64)
	if parseErr == nil {
		return &defaultBlock{blockNumber: decoded}, nil
	}
	// neither
	return nil, fmt.Errorf("not a named block OR failed to parse as a number err %q", parseErr)
}

// integer of a block number, or the string "earliest", "latest" or "pending", as in the default block parameter.
type defaultBlock struct {
	namedBlock  string
	blockNumber uint64
}

func (b *defaultBlock) IsNamedBlock() bool {
	if b.namedBlock == "" {
		return false
	}
	return true
}

// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getblockbynumber
func (s *ethService) GetBlockByNumber(r *http.Request, p *[]interface{}, reply *Block) error {
	fmt.Println("Received a request for GetBlockByNumber")
	params := *p
	fmt.Println("Params are : ", params)

	// handle params
	// must have two params
	numParams := len(params)
	if numParams != 2 {
		return fmt.Errorf("need 2 params, got %q", numParams)
	}
	// first arg is string of block to get
	number, ok := params[0].(string)
	if !ok {
		fmt.Printf("Incorrect argument received: %#v", params[0])
		return fmt.Errorf("Incorrect first parameter sent, must be string")
	}
	block, err := parseAsDefaultBlock(strip0x(number))
	if err != nil {
		return err
	}
	// second arg is bool for full txn or hash txn
	fullTransactions, ok := params[1].(bool)
	if !ok {
		return fmt.Errorf("Incorrect second parameter sent, must be boolean")
	}

	getBlockByNumber := func(number uint64) (Block, error) {
		block, err := s.ledgerClient.QueryBlock(number)
		if err != nil {
			return Block{}, err
		}

		blkHeader := block.GetHeader()

		data := block.GetData().GetData()
		numBlocks := len(data)

		txns := make([]interface{}, numBlocks)

		// drill into the block to find the transaction ids it contains
		for idx, d := range data {
			if d != nil { // can a data be empty? Is this an error?
				env := &common.Envelope{}
				if err := proto.Unmarshal(d, env); err != nil {
					return Block{}, err
				}

				payload := &common.Payload{}
				if err := proto.Unmarshal(env.GetPayload(), payload); err != nil {
					return Block{}, err
				}

				chdr := &common.ChannelHeader{}
				if err := proto.Unmarshal(payload.GetHeader().GetChannelHeader(), chdr); err != nil {
					return Block{}, err
				}

				if fullTransactions {
					fmt.Println("IT WANTS THE FULL TRANSACTIONS")
				} else {
					fmt.Println("IT WANTS JUST THE HASHES")
					txns[idx] = "0x" + chdr.TxId
				}
			}
		}

		blk := Block{
			Number:       "0x" + strconv.FormatUint(number, 16),
			Hash:         "0x" + hex.EncodeToString(blkHeader.GetDataHash()),
			ParentHash:   "0x" + hex.EncodeToString(blkHeader.GetPreviousHash()),
			Transactions: txns,
		}
		fmt.Println("asked for block", number, "found block", blk)
		return blk, nil
	}

	if block.IsNamedBlock() {
		blockName := block.namedBlock
		switch blockName {
		case "latest":
			// latest
			// qscc GetChainInfo, for a BlockchainInfo
			// from that take the height
			// using the height, call GetBlockByNumber

			blockchainInfo, err := s.ledgerClient.QueryInfo()
			if err != nil {
				fmt.Println(err)
				return err
			}

			// height is the block being worked on now, we want the previous block
			topBlockNumber := blockchainInfo.BCI.GetHeight() - 1
			fmt.Println("latest block is", topBlockNumber)
			// handleNumberedBlock topBlockNumber
			*reply, err = getBlockByNumber(topBlockNumber)
			if err != nil {
				fmt.Println(err)
				return err
			}
		case "earliest":
			// handleNumberedBlock 0
			*reply, err = getBlockByNumber(0)
			if err != nil {
				return err
			}
		case "pending":
			// ???
		}
	} else {
		// handleNumberedBlock
		// do we check that it's in bound or go-for-it?
		*reply, err = getBlockByNumber(block.blockNumber)
		if err != nil {
			return fmt.Errorf("")
		}
	}
	return nil
}

type Transaction struct { // object, or null when no transaction was found:
	BlockHash   string `json:"blockHash"`   // DATA, 32 Bytes - hash of the block where this transaction was in. null when its pending.
	BlockNumber string `json:"blockNumber"` // QUANTITY - block number where this transaction was in. null when its pending.
	// from //: DATA, 20 Bytes - address of the sender.
	// gas //: QUANTITY - gas provided by the sender.
	// gasPrice //: QUANTITY - gas price provided by the sender in Wei.
	Hash string `json:"hash"` //: DATA, 32 Bytes - hash of the transaction.
	// input //: DATA - the data send along with the transaction.
	// nonce //: QUANTITY - the number of transactions made by the sender prior to this one.
	// to //: DATA, 20 Bytes - address of the receiver. null when its a contract creation transaction.
	// transactionIndex //: QUANTITY - integer of the transactions index position in the block. null when its pending.
	// value //: QUANTITY - value transferred in Wei.
	// v //: QUANTITY - ECDSA recovery id
	// r //: DATA, 32 Bytes - ECDSA signature r
	// s //: DATA, 32 Bytes - ECDSA signature s
}

// GetTransactionByHash takes a TransactionID as a string and returns
// the details of the transaction.
func (s *ethService) GetTransactionByHash(r *http.Request, txID *string, reply *Transaction) error {
	if *txID == "" {
		return fmt.Errorf("txID was empty")
	}
	strippedTxId := strip0x(*txID)
	fmt.Println("GetTransactionByHash", strippedTxId)

	block, err := s.ledgerClient.QueryBlockByTxID(fab.TransactionID(strippedTxId))
	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err.Error())
	}

	blkHeader := block.GetHeader()

	*reply = Transaction{
		Hash:        *txID,
		BlockHash:   "0x" + hex.EncodeToString(blkHeader.GetDataHash()),
		BlockNumber: "0x" + strconv.FormatUint(blkHeader.GetNumber(), 16),
	}
	return nil
}
