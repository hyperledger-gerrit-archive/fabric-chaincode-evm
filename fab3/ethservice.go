/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab3

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/hyperledger/fabric-chaincode-evm/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var ZeroAddress = make([]byte, 20)

//go:generate counterfeiter -o ../mocks/fab3/mockchannelclient.go --fake-name MockChannelClient ./ ChannelClient

type ChannelClient interface {
	Query(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
	Execute(request channel.Request, options ...channel.RequestOption) (channel.Response, error)
}

//go:generate counterfeiter -o ../mocks/fab3/mockledgerclient.go --fake-name MockLedgerClient ./ LedgerClient

type LedgerClient interface {
	QueryInfo(options ...ledger.RequestOption) (*fab.BlockchainInfoResponse, error)
	QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*common.Block, error)
	QueryBlockByTxID(txid fab.TransactionID, options ...ledger.RequestOption) (*common.Block, error)
	QueryTransaction(txid fab.TransactionID, options ...ledger.RequestOption) (*peer.ProcessedTransaction, error)
}

//go:generate counterfeiter -o ../mocks/fab3/mockethservice.go --fake-name MockEthService ./ EthService

// EthService is the rpc server implementation. Each function is an
// implementation of one ethereum json-rpc
// https://github.com/ethereum/wiki/wiki/JSON-RPC
//
// Arguments and return values are formatted as HEX value encoding
// https://github.com/ethereum/wiki/wiki/JSON-RPC#hex-value-encoding
//
// gorilla RPC is the receiver of these functions, they must all take three
// pointers, and return a single error
//
// see godoc for RegisterService(receiver interface{}, name string) error
//
type EthService interface {
	GetCode(r *http.Request, arg *string, reply *string) error
	Call(r *http.Request, args *EthArgs, reply *string) error
	SendTransaction(r *http.Request, args *EthArgs, reply *string) error
	GetTransactionReceipt(r *http.Request, arg *string, reply *TxReceipt) error
	Accounts(r *http.Request, arg *string, reply *[]string) error
	EstimateGas(r *http.Request, args *EthArgs, reply *string) error
	GetBalance(r *http.Request, p *[]string, reply *string) error
	GetBlockByNumber(r *http.Request, p *[]interface{}, reply *Block) error
	BlockNumber(r *http.Request, _ *interface{}, reply *string) error
	GetTransactionByHash(r *http.Request, txID *string, reply *Transaction) error
	GetLogs(*http.Request, *GetLogsArgs, *[]Log) error
}

type ethService struct {
	channelClient ChannelClient
	ledgerClient  LedgerClient
	channelID     string
	ccid          string
	logger        *zap.SugaredLogger
}

// Incoming structs used as arguments

type EthArgs struct {
	To       string `json:"to"`
	From     string `json:"from"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Data     string `json:"data"`
	Nonce    string `json:"nonce"`
}

type GetLogsArgs struct {
	FromBlock string
	ToBlock   string
	Address   addressFilter
	Topics    topicsFilter
}

func (gla *GetLogsArgs) UnmarshalJSON(data []byte) error {
	type inputGetLogsArgs struct {
		FromBlock string      `json:"fromBlock"`
		ToBlock   string      `json:"toBlock"`
		Address   interface{} `json:"address"` // string or array of strings.
		Topics    interface{} `json:"topics"`  // array of strings, or array of array of strings
	}
	var input inputGetLogsArgs
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}
	gla.FromBlock = input.FromBlock
	gla.ToBlock = input.ToBlock

	var af addressFilter
	// handle the address(es)
	// zap.S().Debug("addresses", input.Address)
	// DATA|Array, 20 Bytes - (optional) Contract address or a list of
	// addresses from which logs should originate.
	if singleAddress, ok := input.Address.(string); ok {
		a, err := NewAddressFilter(singleAddress)
		if err != nil {
			return err
		}
		af = append(af, a...)
	} else if multipleAddresses, ok := input.Address.([]interface{}); ok {
		for _, address := range multipleAddresses {
			if singleAddress, ok := address.(string); ok {
				a, err := NewAddressFilter(singleAddress)
				if err != nil {
					return err
				}
				af = append(af, a...)
			}
		}
	} else {
		return fmt.Errorf("badly formatted address field")
	}

	gla.Address = af

	var tf topicsFilter

	// handle the topics parsing
	// need to strip 0x prefix, as stored
	zap.S().Debug("topics", input.Topics)
	if topics, ok := input.Topics.([]interface{}); !ok {
		return fmt.Errorf("topics must be slice")
	} else {
		for i, topic := range topics {
			if singleTopic, ok := topic.(string); ok {
				zap.S().Debug("single topic", "index", i)
				f, err := NewTopicFilter(singleTopic)
				if err != nil {
					return err
				}
				tf = append(tf, f)
			} else if multipleTopic, ok := topic.([]interface{}); ok {
				zap.S().Debug("or'd topics", "index", i)
				var mtf topicFilter
				for _, singleTopic := range multipleTopic {
					if stringTopic, ok := singleTopic.(string); ok {
						f, err := NewTopicFilter(stringTopic)
						if err != nil {
							return err
						}
						mtf = append(mtf, f...)
					} else {
						return fmt.Errorf("all topics must be strings")
					}
				}
				tf = append(tf, mtf)
			} else {
				return fmt.Errorf("some unparsable trash %q", topic)
			}
		}
	}

	gla.Topics = tf

	return nil
}

// structs being returned

type TxReceipt struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionIndex  string `json:"transactionIndex"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       string `json:"blockNumber"`
	ContractAddress   string `json:"contractAddress"`
	GasUsed           int    `json:"gasUsed"`
	CumulativeGasUsed int    `json:"cumulativeGasUsed"`
	To                string `json:"to"`
	Logs              []Log  `json:"logs"`
	Status            string `json:"status"`
}

type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data,omitempty"`
	BlockNumber string   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     string   `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	Index       string   `json:"logIndex"`
}

// Transaction represents an ethereum evm transaction.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#returns-28
type Transaction struct { // object, or null when no transaction was found:
	BlockHash   string `json:"blockHash"`   // DATA, 32 Bytes - hash of the block where this transaction was in. null when its pending.
	BlockNumber string `json:"blockNumber"` // QUANTITY - block number where this transaction was in. null when its pending.
	To          string `json:"to"`          // DATA, 20 Bytes - address of the receiver. null when its a contract creation transaction.
	// From is generated by EVM Chaincode. Until account generation
	// stabilizes, we are not returning a value.
	//
	// From can be gotten from the Signature on the Transaction Envelope
	//
	// From string `json:"from"` // DATA, 20 Bytes - address of the sender.
	Input            string `json:"input"`            // DATA - the data send along with the transaction.
	TransactionIndex string `json:"transactionIndex"` // QUANTITY - integer of the transactions index position in the block. null when its pending.
	Hash             string `json:"hash"`             //: DATA, 32 Bytes - hash of the transaction.
}

// Block is an eth return struct
// defined https://github.com/ethereum/wiki/wiki/JSON-RPC#returns-26
type Block struct {
	Number     string `json:"number"`     // number: QUANTITY - the block number. null when its pending block.
	Hash       string `json:"hash"`       // hash: DATA, 32 Bytes - hash of the block. null when its pending block.
	ParentHash string `json:"parentHash"` // parentHash: DATA, 32 Bytes - hash of the parent block.
	// size: QUANTITY - integer the size of this block in bytes.
	// timestamp: QUANTITY - the unix timestamp for when the block was collated.
	Transactions []interface{} `json:"transactions"` // transactions: Array - Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter.
}

func NewEthService(channelClient ChannelClient, ledgerClient LedgerClient, channelID string, ccid string, logger *zap.SugaredLogger) EthService {
	return &ethService{channelClient: channelClient, ledgerClient: ledgerClient, channelID: channelID, ccid: ccid, logger: logger.Named("ethservice")}
}

func (s *ethService) GetCode(r *http.Request, arg *string, reply *string) error {
	strippedAddr := strip0x(*arg)

	response, err := s.query(s.ccid, "getCode", [][]byte{[]byte(strippedAddr)})

	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err)
	}

	*reply = string(response.Payload)

	return nil
}

func (s *ethService) Call(r *http.Request, args *EthArgs, reply *string) error {
	response, err := s.query(s.ccid, strip0x(args.To), [][]byte{[]byte(strip0x(args.Data))})

	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err)
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
		return fmt.Errorf("Failed to execute transaction: %s", err)
	}
	*reply = string(response.TransactionID)
	return nil
}

func (s *ethService) GetTransactionReceipt(r *http.Request, txID *string, reply *TxReceipt) error {
	strippedTxID := strip0x(*txID)

	block, err := s.ledgerClient.QueryBlockByTxID(fab.TransactionID(strippedTxID))
	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err)
	}

	blkHeader := block.GetHeader()

	transactionsFilter := block.GetMetadata().GetMetadata()[common.BlockMetadataIndex_TRANSACTIONS_FILTER]

	receipt := TxReceipt{
		TransactionHash:   "0x" + strippedTxID,
		BlockHash:         "0x" + hex.EncodeToString(blkHeader.GetDataHash()),
		BlockNumber:       "0x" + strconv.FormatUint(blkHeader.GetNumber(), 16),
		GasUsed:           0,
		CumulativeGasUsed: 0,
	}

	index, txPayload, err := findTransaction(strippedTxID, block.GetData().GetData())
	if err != nil {
		return fmt.Errorf("Failed parsing the transactions in the block: %s", err)
	}

	receipt.TransactionIndex = index
	indexU, _ := strconv.ParseUint(strip0x(index), 16, 64)
	// for fabric transactions, 0 is valid, 1 is invalid, the opposite of how ethereum
	receipt.Status = "0x" + strconv.FormatUint(((1+uint64(transactionsFilter[indexU]))%2), 16)

	to, _, respPayload, err := getTransactionInformation(txPayload)

	if to != "" {
		callee, err := hex.DecodeString(to)
		if err != nil {
			return fmt.Errorf("Failed to decode to address: %s", err)
		}

		if bytes.Equal(callee, ZeroAddress) {
			receipt.ContractAddress = string(respPayload.GetResponse().GetPayload())
		} else {
			receipt.To = "0x" + to
		}
	}

	txLogs, err := fabricEventToEVMLogs(respPayload.Events, receipt.BlockNumber, receipt.TransactionHash, receipt.TransactionIndex, receipt.BlockHash, nil, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get EVM Logs out of fabric event")
	}
	receipt.Logs = txLogs

	*reply = receipt
	return nil
}

func (s *ethService) Accounts(r *http.Request, arg *string, reply *[]string) error {
	response, err := s.query(s.ccid, "account", [][]byte{})
	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err)
	}

	*reply = []string{"0x" + strings.ToLower(string(response.Payload))}

	return nil
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
	s.logger.Debug("EstimateGas called")
	*reply = "0x0"
	return nil
}

// GetBalance takes an address and a block, but this implementation
// does not check or use either parameter.
//
// Always returns zero.
func (s *ethService) GetBalance(r *http.Request, p *[]string, reply *string) error {
	s.logger.Debug("GetBalance called")
	*reply = "0x0"
	return nil
}

// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getblockbynumber
func (s *ethService) GetBlockByNumber(r *http.Request, p *[]interface{}, reply *Block) error {
	s.logger.Debug("Received a request for GetBlockByNumber")
	params := *p
	s.logger.Debug("Params are : ", params)

	// handle params
	// must have two params
	numParams := len(params)
	if numParams != 2 {
		return fmt.Errorf("need 2 params, got %q", numParams)
	}
	// first arg is string of block to get
	number, ok := params[0].(string)
	if !ok {
		s.logger.Debugf("Incorrect argument received: %#v", params[0])
		return fmt.Errorf("Incorrect first parameter sent, must be string")
	}

	// second arg is bool for full txn or hash txn
	fullTransactions, ok := params[1].(bool)
	if !ok {
		return fmt.Errorf("Incorrect second parameter sent, must be boolean")
	}

	parsedNumber, err := s.parseBlockNum(strip0x(number))
	if err != nil {
		return err
	}

	block, err := s.ledgerClient.QueryBlock(parsedNumber)
	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %v", err)
	}

	blkHeader := block.GetHeader()

	blockHash := "0x" + hex.EncodeToString(blkHeader.GetDataHash())
	blockNumber := "0x" + strconv.FormatUint(parsedNumber, 16)

	// each data is a txn
	data := block.GetData().GetData()
	txns := make([]interface{}, len(data))

	// drill into the block to find the transaction ids it contains
	for index, transactionData := range data {
		if transactionData == nil {
			continue
		}

		payload, chdr, err := getChannelHeaderandPayloadFromBlockdata(transactionData)
		if err != nil {
			return err
		}

		// returning full transactions is unimplemented,
		// so the hash-only case is the only case.
		s.logger.Debug("block has transaction hash:", chdr.TxId)

		if fullTransactions {
			txn := Transaction{
				BlockHash:        blockHash,
				BlockNumber:      blockNumber,
				TransactionIndex: "0x" + strconv.FormatUint(uint64(index), 16),
				Hash:             "0x" + chdr.TxId,
			}
			to, input, _, err := getTransactionInformation(payload)
			if err != nil {
				return err
			}

			txn.To = "0x" + to
			txn.Input = "0x" + input
			txns[index] = txn
		} else {
			txns[index] = "0x" + chdr.TxId
		}
	}

	blk := Block{
		Number:       blockNumber,
		Hash:         blockHash,
		ParentHash:   "0x" + hex.EncodeToString(blkHeader.GetPreviousHash()),
		Transactions: txns,
	}
	s.logger.Debug("asked for block", number, "found block", blk)

	*reply = blk
	return nil
}

func (s *ethService) BlockNumber(r *http.Request, _ *interface{}, reply *string) error {
	blockNumber, err := s.parseBlockNum("latest")
	if err != nil {
		return fmt.Errorf("failed to get latest block number: %s", err)
	}
	*reply = "0x" + strconv.FormatUint(uint64(blockNumber), 16)

	return nil
}

// GetTransactionByHash takes a TransactionID as a string and returns the
// details of the transaction.
//
// The implementation of this function follows the EVM ChainCode implementation
// of Invoke.
//
// Since this takes only one string, we can have gorilla verify that it has
// received only a single string, and skip our own verification.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_gettransactionbyhash
func (s *ethService) GetTransactionByHash(r *http.Request, txID *string, reply *Transaction) error {
	strippedTxId := strip0x(*txID)

	if strippedTxId == "" {
		return fmt.Errorf("txID was empty")
	}
	s.logger.Debug("GetTransactionByHash", strippedTxId) // logging input to function

	txn := Transaction{
		Hash: "0x" + strippedTxId,
	}

	block, err := s.ledgerClient.QueryBlockByTxID(fab.TransactionID(strippedTxId))
	if err != nil {
		return fmt.Errorf("Failed to query the ledger: %s", err)
	}
	blkHeader := block.GetHeader()
	txn.BlockHash = "0x" + hex.EncodeToString(blkHeader.GetDataHash())
	txn.BlockNumber = "0x" + strconv.FormatUint(blkHeader.GetNumber(), 16)

	index, txPayload, err := findTransaction(strippedTxId, block.GetData().GetData())
	if err != nil {
		return fmt.Errorf("Failed to parse through transactions in the block: %s", err)
	}

	txn.TransactionIndex = index

	to, input, _, err := getTransactionInformation(txPayload)
	if err != nil {
		return err
	}

	if to != "" {
		txn.To = "0x" + to
	}

	if input != "" {
		txn.Input = "0x" + input
	}

	*reply = txn
	return nil
}

type addressFilter []string // 20 Byte Addresses

// NewAddressFilter takes a string and checks that is the correct length to
// represent a topic and strips the 0x
func NewAddressFilter(s string) (addressFilter, error) {
	if len(s) != HexEncodedAddressLegnth {
		return nil, fmt.Errorf("address in wrong format, not 42 chars %q", s)
	}
	return addressFilter{strip0x(s)}, nil
}

type topicFilter []string // 32 Byte Topics

// NewTopicFilter takes a string and checks that is the correct length to
// represent a topic and strips the 0x
func NewTopicFilter(s string) (topicFilter, error) {
	if len(s) != HexEncodedTopicLegnth {
		return nil, fmt.Errorf("topic in wrong format, %d chars not 66 chars %q", len(s), s)
	}
	return topicFilter{strip0x(s)}, nil
}

// slices are in order, but maybe we should name individual fields first,second,
// etc.  only a few individual fields.
//
// TODO(mhb): maybe an alias for the contained type instead of a struct?
type topicsFilter []topicFilter

func NewTopicsFilter(tf ...topicFilter) topicsFilter {
	return tf
}

const (
	HexEncodedAddressLegnth = 42 // 20 bytes, is 40 hex chars, plus two for '0x'
	HexEncodedTopicLegnth   = 66 // 32 bytes, is 64 hex chars, plus two for '0x'
)

//GetLogs currently returns all logs in range FromBlock to ToBlock
func (s *ethService) GetLogs(r *http.Request, args *GetLogsArgs, logs *[]Log) error {
	s.logger.Debug("GetLogs called")
	s.logger.Debug("params are", args)

	// check for earliest
	if args.FromBlock == "earliest" || args.ToBlock == "earliest" {
		return fmt.Errorf("unsupported: fabric does not have the concept of in-progress blocks being visible.")
	}
	// set defaults *after* checking for input conflicts and validating
	if args.FromBlock == "" {
		args.FromBlock = "latest"
	}
	if args.ToBlock == "" {
		args.ToBlock = "latest"
	}

	// function to create filter object from addresses and topics?
	var af addressFilter
	af = args.Address
	var tf topicsFilter
	tf = args.Topics

	var from, to uint64

	// maybe check if both from and to are 'latest' to avoid doing a query
	// twice and coming out with different answers each time, which doesn't
	// hurt, but is weird.
	from, err := s.parseBlockNum(strip0x(args.FromBlock))
	if err != nil {
		return errors.Wrap(err, "failed to parse the block number")
	}
	// check if both from and to are the same to avoid doing two
	// queries to the fabric network.
	if args.FromBlock == args.ToBlock {
		to = from
	} else {
		to, err = s.parseBlockNum(strip0x(args.ToBlock))
		if err != nil {
			return errors.Wrap(err, "failed to parse the block number")
		}
	}

	if from > to {
		return fmt.Errorf("fromBlock number greater than toBlock number")
	}

	var txLogs []Log

	s.logger.Debugw("handling blocks", "from", from, "to", to)
	for blockNumber := from; blockNumber <= to; blockNumber++ {
		s.logger.Debug("Block", blockNumber)
		block, err := s.ledgerClient.QueryBlock(blockNumber)
		if err != nil {
			return fmt.Errorf("failed to query the ledger: %v", err)
		}
		blockHeader := block.GetHeader()
		blockHash := "0x" + hex.EncodeToString(blockHeader.GetDataHash())
		blockData := block.GetData().GetData()
		transactionsFilter := block.GetMetadata().GetMetadata()[common.BlockMetadataIndex_TRANSACTIONS_FILTER]
		s.logger.Debug("handling ", len(blockData), " transactions in block")
		for transactionIndex, transactionData := range blockData {
			// check validity of transaction
			if (transactionsFilter[transactionIndex] == 1) || (transactionData == nil) {
				continue
			}

			// start processing the transaction
			payload, chdr, err := getChannelHeaderandPayloadFromBlockdata(transactionData)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal the transaction")
			}

			transactionHash := "0x" + chdr.TxId
			s.logger.Debug("transaction ", transactionIndex, " has hash ", transactionHash)

			var respPayload *peer.ChaincodeAction
			_, _, respPayload, err = getTransactionInformation(payload)
			if err != nil {
				return errors.Wrap(err, "failed to unmarshal the transaction details")
			}

			s.logger.Debug("transaction payload", respPayload)

			blkNumber := "0x" + strconv.FormatUint(blockNumber, 16)

			logs, err := fabricEventToEVMLogs(respPayload.Events, blkNumber, transactionHash, "rast", blockHash, af, tf)
			if err != nil {
				return errors.Wrap(err, "failed to get EVM Logs out of fabric event")
			}
			txLogs = append(txLogs, logs...)
		}
	}

	s.logger.Debug("returning logs", txLogs)
	*logs = txLogs

	return nil
}

func getChannelHeaderandPayloadFromBlockdata(transactionData []byte) (*common.Payload, *common.ChannelHeader, error) {
	env := &common.Envelope{}
	if err := proto.Unmarshal(transactionData, env); err != nil {
		return nil, nil, err
	}

	payload := &common.Payload{}
	if err := proto.Unmarshal(env.GetPayload(), payload); err != nil {
		return nil, nil, err
	}
	chdr := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.GetHeader().GetChannelHeader(), chdr); err != nil {
		return nil, nil, err
	}
	return payload, chdr, nil
}

func (s *ethService) query(ccid, function string, queryArgs [][]byte) (channel.Response, error) {
	return s.channelClient.Query(channel.Request{
		ChaincodeID: ccid,
		Fcn:         function,
		Args:        queryArgs,
	})
}

// https://github.com/ethereum/wiki/wiki/JSON-RPC#the-default-block-parameter
func (s *ethService) parseBlockNum(input string) (uint64, error) {
	// check if it's one of the named-blocks
	switch input {
	case "latest":
		// latest
		// qscc GetChainInfo, for a BlockchainInfo
		// from that take the height
		// using the height, call GetBlockByNumber

		blockchainInfo, err := s.ledgerClient.QueryInfo()
		if err != nil {
			s.logger.Debug(err)
			return 0, fmt.Errorf("failed to query the ledger: %v", err)
		}

		// height is the block being worked on now, we want the previous block
		topBlockNumber := blockchainInfo.BCI.GetHeight() - 1
		return topBlockNumber, nil
	case "earliest":
		return 0, nil
	case "pending":
		return 0, fmt.Errorf("unsupported: fabric does not have the concept of in-progress blocks being visible")
	default:
		return strconv.ParseUint(input, 16, 64)
	}
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

// getTransactionInformation takes a payload
// It returns if available the To, Input, the Response Payload of the transaction in the payload, otherwise it returns an error
func getTransactionInformation(payload *common.Payload) (string, string, *peer.ChaincodeAction, error) {
	txActions := &peer.Transaction{}
	err := proto.Unmarshal(payload.GetData(), txActions)
	if err != nil {
		return "", "", nil, err
	}

	ccPropPayload, respPayload, err := getPayloads(txActions.GetActions()[0])
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to unmarshal transaction: %s", err)
	}

	invokeSpec := &peer.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(ccPropPayload.GetInput(), invokeSpec)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to unmarshal transaction: %s", err)
	}

	// callee, input data is standard case, also handle getcode & account cases
	args := invokeSpec.GetChaincodeSpec().GetInput().Args

	if len(args) != 2 || string(args[0]) == "getCode" {
		// no more data available to fill the transaction
		return "", "", respPayload, nil
	}

	// At this point, this is either an EVM Contract Deploy,
	// or an EVM Contract Invoke. We don't care about the
	// specific case, fill in the fields directly.

	// First arg is to and second arg is the input data
	return string(args[0]), string(args[1]), respPayload, nil
}

// findTransaction takes in the txId and  block data from block.GetData().GetData() where block is of type *common.Block
// It returns the index of the transaction, transaction payload, otherwise it returns an error
func findTransaction(txID string, blockData [][]byte) (string, *common.Payload, error) {
	for index, transactionData := range blockData {
		if transactionData == nil { // can a data be empty? Is this an error?
			continue
		}

		payload, chdr, err := getChannelHeaderandPayloadFromBlockdata(transactionData)
		if err != nil {
			return "", &common.Payload{}, err
		}

		// early exit to try next transaction
		if txID != chdr.TxId {
			// transaction does not match, go to next
			continue
		}

		return "0x" + strconv.FormatUint(uint64(index), 16), payload, nil
	}

	return "", &common.Payload{}, nil
}

func fabricEventToEVMLogs(events []byte, blocknumber, txhash, txindex, blockhash string, af addressFilter, tf topicsFilter) ([]Log, error) {
	if events == nil {
		return nil, nil
	}

	chaincodeEvent := &peer.ChaincodeEvent{}
	err := proto.Unmarshal(events, chaincodeEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode chaincode event: %s", err)
	}

	var eventMsgs []event.Event
	err = json.Unmarshal(chaincodeEvent.Payload, &eventMsgs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal chaincode event payload: %s", err)
	}

	var txLogs []Log
	for i, logEvent := range eventMsgs {

		zap.S().Debug("checking event for matching address")
		if af != nil {
			foundMatch := false
			for _, address := range af { // if no address, empty range, skipped
				zap.S().Debug("matching address", "matcherAddress", address, "eventAddress", logEvent.Address)
				if logEvent.Address == address {
					foundMatch = true
					break
				}
			}
			if !foundMatch {
				continue // no match, move to next logEvent
			}
		}

		zap.S().Debug("checking for topics")
		allMatch := true // opposite is any not match
		// check match for each topic,
		for i, topicFilter := range tf {
			// if filter is empty it matches automatically.
			if len(topicFilter) == 0 {
				continue
			}

			eventTopic := logEvent.Topics[i]
			foundMatch := false
			for _, topic := range topicFilter {
				zap.S().Debug("matching Topic ", "matcherTopic", topic, "eventTopic", eventTopic)
				if topic == eventTopic || topic == "" {
					foundMatch = true
					break
				}
			}
			if foundMatch == false {
				allMatch = false
				// if we didn't find a match, no use in checking any of the other topics
				break
			}

		}
		if allMatch == false {
			continue
		}

		topics := []string{}
		for _, topic := range logEvent.Topics {
			topics = append(topics, "0x"+topic)
		}
		logObj := Log{
			Address:     "0x" + logEvent.Address,
			Topics:      topics,
			BlockNumber: blocknumber,
			TxHash:      txhash,
			TxIndex:     txindex,
			BlockHash:   blockhash,
			Index:       "0x" + strconv.FormatUint(uint64(i), 16),
		}

		if logEvent.Data != "" {
			logObj.Data = "0x" + logEvent.Data
		}

		txLogs = append(txLogs, logObj)
	}
	return txLogs, nil
}
