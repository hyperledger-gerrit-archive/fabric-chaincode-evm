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
	"reflect"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
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
	//QueryBlockByHash(blockHash []byte, options ...ledger.RequestOption) (*common.Block, error)
}

// EthService is the rpc server implementation. Each function is an
// implementation of one ethereum json-rpc
// https://github.com/ethereum/wiki/wiki/JSON-RPC
//
// Arguments and return values are formatted as HEX value encoding
// https://github.com/ethereum/wiki/wiki/JSON-RPC#hex-value-encoding
//
//go:generate counterfeiter -o ../mocks/fab3/mockethservice.go --fake-name MockEthService ./ EthService
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

// consider implementing json.unmarshal and using custom types for address and topics
type GetLogsArgs struct {
	FromBlock string `json:"fromBlock,omitempty"`
	// QUANTITY|TAG - (optional, default: "latest") Integer block number, or
	// "latest" for the last mined block or "pending", "earliest" for not
	// yet mined transactions.
	ToBlock string `json:"toBlock,omitempty"`
	// QUANTITY|TAG - (optional, default: "latest") Integer block number, or
	// "latest" for the last mined block or "pending", "earliest" for not
	// yet mined transactions.
	Address interface{} `json:"address,omitempty"` // string or array of strings.
	// DATA|Array, 20 Bytes - (optional) Contract address or a list of
	// addresses from which logs should originate.
	Topics interface{} `json:"topics,omitempty"` // array of strings or array of array of strings
	// Array of DATA, - (optional) Array of 32 Bytes DATA topics. Topics are
	// order-dependent. Each topic can also be an array of DATA with "or"
	// options.
	Blockhash string `json:"blockhash,omitempty"`
	// DATA, 32 Bytes (optional) restricts the logs returned to the single
	// block with the 32-byte hash blockHash. Using blockHash is equivalent
	// to fromBlock = toBlock = the block number with hash blockHash. If
	// blockHash is present in the filter criteria, then neither fromBlock
	// nor toBlock are allowed.
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
	Data        string   `json:"data"`
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

	if respPayload.Events != nil {
		chaincodeEvent, err := getChaincodeEvents(respPayload)
		if err != nil {
			return fmt.Errorf("Failed to decode chaincode event: %s", err)
		}

		var eventMsgs []event.Event
		err = json.Unmarshal(chaincodeEvent.Payload, &eventMsgs)
		if err != nil {
			s.logger.Info(chaincodeEvent.Payload)
			return fmt.Errorf("Failed to unmarshal chaincode event payload: %s", err)
		}

		var txLogs []Log
		txLogs = make([]Log, 0)
		for i, logEvent := range eventMsgs {
			topics := []string{}
			for _, topic := range logEvent.Topics {
				topics = append(topics, "0x"+topic)
			}
			logObj := Log{
				Address:     "0x" + logEvent.Address,
				Topics:      topics,
				Data:        "0x" + logEvent.Data,
				BlockNumber: receipt.BlockNumber,
				TxHash:      receipt.TransactionHash,
				TxIndex:     receipt.TransactionIndex,
				BlockHash:   "0x" + hex.EncodeToString(blkHeader.GetDataHash()),
				Index:       "0x" + strconv.FormatUint(uint64(i), 16),
			}
			txLogs = append(txLogs, logObj)
		}
		receipt.Logs = txLogs
	} else {
		receipt.Logs = nil
	}

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
		env := &common.Envelope{}
		if err := proto.Unmarshal(transactionData, env); err != nil {
			return err
		}

		payload := &common.Payload{}
		if err := proto.Unmarshal(env.GetPayload(), payload); err != nil {
			return err
		}

		chdr := &common.ChannelHeader{}
		if err := proto.Unmarshal(payload.GetHeader().GetChannelHeader(), chdr); err != nil {
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

// slices are in order, but maybe we should name individual fields first,second,
// etc.  only a few individual fields.
//
// TODO(mhb): maybe an alias for the contained type instead of a struct?
type topicFilter struct {
	// topics to be or'd together
	topics []string
}

type logFilter struct {
	addresses []string
	topics    []topicFilter
}

const (
	HexEncodedAddressLegnth = 42 // 20 bytes, is 40 hex chars, plus two for '0x'
	HexEncodedTopicData     = 66 // 32 bytes, is 64 hex chars, plus two for '0x'
)

/* GetLogs

try to get all local calculations done before we make any calls to remote system

currently returns all logs in range with no filtering at all
*/
func (s *ethService) GetLogs(r *http.Request, args *GetLogsArgs, reply *[]Log) error {
	s.logger.Debug(args)

	if (args.FromBlock != "" || args.ToBlock != "") && args.Blockhash != "" {
		return fmt.Errorf("Blockhash cannot be set at the same time as From or To; blockhash %q, From %q, To %q", args.Blockhash, args.FromBlock, args.ToBlock)
	}
	// check for earliest
	if args.FromBlock == "earliest" || args.ToBlock == "earliest" {
		return fmt.Errorf("Unimplemented: fabric does not have the concept of in-progress blocks being visible.")
	}
	// set defaults *after* checking for input conflicts and validating
	if args.FromBlock == "" {
		args.FromBlock = "latest"
	}
	if args.ToBlock == "" {
		args.ToBlock = "latest"
	}

	// function to create filter object from addresses and topics?
	var lf logFilter
	// handle the address(es)
	// need to strip 0x prefix, as stored
	s.logger.Debug("addresses", args.Address)
	if singleAddress, ok := args.Address.(string); ok {
		if len(singleAddress) != HexEncodedAddressLegnth {
			return fmt.Errorf("address in wrong format, not 42 chars %q", singleAddress)
		}
		lf.addresses = append(lf.addresses, singleAddress)
	} else if multipleAddresses, ok := args.Address.([]string); ok {
		for _, address := range multipleAddresses {
			if len(address) != HexEncodedAddressLegnth {
				return fmt.Errorf("address in wrong format, not 42 chars %q", address)
			}
		}
		lf.addresses = multipleAddresses
	}

	// handle the topics parsing
	// need to strip 0x prefix, as stored
	if args.Topics != nil {
		//rt := reflect.TypeOf(args.Topics)
		s.logger.Debug("topics", args.Topics)
		// should always be a slice
		//if rt.Kind() != reflect.Slice {
		//	return fmt.Errorf("topics must be slice")
		//}

		if topics, ok := args.Topics.([]interface{}); !ok {
			return fmt.Errorf("topics must be slice")
		} else {
			for i, topic := range topics {
				rt := reflect.TypeOf(topic)
				s.logger.Debug(topic, rt, rt.Kind())
				if singleTopic, ok := topic.(string); ok {
					s.logger.Debug("single topic", "index", i)
					lf.topics = append(lf.topics, topicFilter{[]string{strip0x(singleTopic)}})
				} else if multipleTopic, ok := topic.([]interface{}); ok {
					s.logger.Debug("or'd topics", "index", i)
					var tf topicFilter
					for _, singleTopic := range multipleTopic {
						if stringTopic, ok := singleTopic.(string); ok {
							tf.topics = append(tf.topics, strip0x(stringTopic))
						} else {
							return fmt.Errorf("all topics must be strings")
						}
					}
					lf.topics = append(lf.topics, tf)
				} else {
					return fmt.Errorf("some unparsable trash", topic)
				}
			}
		}
		// create topic filter
		// filter.AddTopicFilter
		//var topicSlice []interface{}

		//have interface{}
		//
		// should be a slice of something
		//
		// slice of strings
		//
		// slice of slices

		/*
			for i, topic := range topicSlice {
			if singleTopic, ok := topic.(string); ok {
					s.logger.Debug("single topic", "index", i)
					lf.topics = append(lf.topics, topicFilter{[]string{singleTopic}})
				} else if multipleTopic, ok := topic.([]string); ok {
					s.logger.Debug("or'd topics", "index", i)
					lf.topics = append(lf.topics, topicFilter{multipleTopic})
				}
			}
		*/
	}
	s.logger.Debug("Final Filter", lf)

	/* OPTIONS

		   1. make filter and apply while iterating

		   2. accumulate all the logs and filter after.

	           3. don't do anything until we know we have to filter, then parse the
	           filters there (probably better to fail fast on a non-appropriate
	           filter.

		   Z. cache everything we see while iterating
	*/

	/* if blockhash is set, this is a single block being parsed

		 TODO prep blockhash for sending to fabric

	         or do a query lookup of what blockNumber it is, and then a common case
	         of handleBlock, same as below.  */
	if args.Blockhash != "" {
		// block, err := s.ledgerClient.QueryBlockByHash(args.Blockhash)
		// if err != nil {
		// 	return fmt.Errorf("Failed to query the ledger: %v", err)
		// }
		// TODO handle the block
		return nil
	}

	// maybe check if both from and to are 'latest' to avoid doing a query
	// twice and coming out with different answers each time, which doesn't
	// hurt, but is weird.

	from, err := s.parseBlockNum(strip0x(args.FromBlock))
	if err != nil {
		return err
	}
	to, err := s.parseBlockNum(strip0x(args.ToBlock))
	if err != nil {
		return err
	}
	if from > to {
		return fmt.Errorf("first block number greater than last block number")
	}

	var txLogs []Log
	//txLogs = make([]Log, 0)

	// possibly multiple blocks in a range
	// TODO start handling each individual block in a loop
	s.logger.Debug("handling blocks", "from", from, "to", to)
	for blockNumber := from; blockNumber <= to; blockNumber++ {
		s.logger.Debug("block", blockNumber)
		// TODO handle the block
		//
		// func getFilteredLogsInBlock(blockNumber, filter)
		block, err := s.ledgerClient.QueryBlock(blockNumber)
		if err != nil {
			return fmt.Errorf("Failed to query the ledger: %v", err)
		}
		blockHeader := block.GetHeader()
		blockHash := "0x" + hex.EncodeToString(blockHeader.GetDataHash())
		blockData := block.GetData().GetData()
		s.logger.Debug("handling ", len(blockData), " transactions in block")
		for transactionIndex, transactionData := range blockData {
			if transactionData == nil { // can a data be empty? Is this an error?
				continue
			}
			env := &common.Envelope{}
			if err := proto.Unmarshal(transactionData, env); err != nil {
				return err
			}

			payload := &common.Payload{}
			if err := proto.Unmarshal(env.GetPayload(), payload); err != nil {
				return err
			}
			chdr := &common.ChannelHeader{}
			if err := proto.Unmarshal(payload.GetHeader().GetChannelHeader(), chdr); err != nil {
				return err
			}

			transactionHash := chdr.TxId
			s.logger.Debug("transaction ", transactionIndex, " has hash ", transactionHash)
			// return of findtransaction equiv

			// alternate
			var respPayload *peer.ChaincodeAction
			_, _, respPayload, err = getTransactionInformation(payload)
			// end alternate

			/* original impl
			txActions := &peer.Transaction{}
			err := proto.Unmarshal(payload.GetData(), txActions)
			if err != nil {
				return err
			}
			// check validness of transaction and abandon
			// if notValid {
			// continue;
			// }

			_, respPayload, err = getPayloads(txActions.GetActions()[0])
			if err != nil {
				s.logger.Debug(txActions.GetActions())
				return fmt.Errorf("Failed to unmarshal transaction: %s", err)
			}
			*/
			s.logger.Debug("transaction payload", respPayload)
			if respPayload.Events != nil {
				chaincodeEvent, err := getChaincodeEvents(respPayload)
				if err != nil {
					return fmt.Errorf("Failed to decode chaincode event: %s", err)
				}

				//Begin common code as "pull all logs out of txn"
				//
				// could format as "pull filtered logs out, with
				// an optional filter, and default to "all logs"
				// and reuse same code in existing areas.
				var eventMsgs []event.Event
				err = json.Unmarshal(chaincodeEvent.Payload, &eventMsgs)
				if err != nil {
					return fmt.Errorf("Failed to unmarshal chaincode event payload: %s", err)
				}

				s.logger.Debug("checking events", eventMsgs)

				for i, logEvent := range eventMsgs {
					s.logger.Debug("checking event", i, logEvent)
					if len(lf.addresses) > 0 {
						s.logger.Debug("checking event for matching dddress")
						foundMatch := false
						for _, address := range lf.addresses { // if no address, empty range, skipped
							s.logger.Debug("matching address", "matcherAddress", address, "eventAddress", logEvent.Address)
							if logEvent.Address == address {
								foundMatch = true
								break
							}
						}
						if foundMatch == false {
							continue // no match, move to next logEvent
						}
					}

					// if there are fewer topics in the event than topics in the filter, it cannot match
					if len(logEvent.Topics) < len(lf.topics) {
						s.logger.Debug("skipping event because topics have different lengths and can never match")
						continue
					}

					s.logger.Debug("checking for topics")
					allMatch := true // opposite is any not match
					// check match for each topic,
					for i, topicFilter := range lf.topics {
						// if filter is empty it matches automatically.
						if len(topicFilter.topics) == 0 {
							continue
						}

						eventTopic := logEvent.Topics[i]
						foundMatch := false
						for _, topic := range topicFilter.topics {
							s.logger.Debug("matching Topic ", "matcherTopic", topic, "eventTopic", eventTopic)
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

					// everything matches, construct the log to return
					topics := []string{}
					for _, topic := range logEvent.Topics {
						// each topic is a hexencoded word256
						topics = append(topics, "0x"+topic)
					}
					logObj := Log{
						Address:     "0x" + logEvent.Address,
						Topics:      topics,
						Data:        "0x" + logEvent.Data,
						BlockNumber: "0x" + strconv.FormatUint(blockNumber, 16),
						TxHash:      transactionHash,
						TxIndex:     "0x" + strconv.FormatUint(uint64(transactionIndex), 16),
						BlockHash:   blockHash,
						Index:       "0x" + strconv.FormatUint(uint64(i), 16),
					}
					txLogs = append(txLogs, logObj)
				}
			}
		}
	}

	s.logger.Debug("returning logs", txLogs)

	s.logger.Debug("logs before being set", reply)
	*reply = txLogs

	return nil
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
			return 0, fmt.Errorf("Failed to query the ledger: %v", err)
		}

		// height is the block being worked on now, we want the previous block
		topBlockNumber := blockchainInfo.BCI.GetHeight() - 1
		return topBlockNumber, nil
	case "earliest":
		return 0, nil
	case "pending":
		return 0, fmt.Errorf("Unimplemented: fabric does not have the concept of in-progress blocks being visible.")
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
		env := &common.Envelope{}
		if err := proto.Unmarshal(transactionData, env); err != nil {
			return "", &common.Payload{}, err
		}

		payload := &common.Payload{}
		if err := proto.Unmarshal(env.GetPayload(), payload); err != nil {
			return "", &common.Payload{}, err
		}

		chdr := &common.ChannelHeader{}
		if err := proto.Unmarshal(payload.GetHeader().GetChannelHeader(), chdr); err != nil {
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

func getChaincodeEvents(respPayload *peer.ChaincodeAction) (*peer.ChaincodeEvent, error) {
	eBytes := respPayload.Events
	chaincodeEvent := &peer.ChaincodeEvent{}
	err := proto.Unmarshal(eBytes, chaincodeEvent)
	return chaincodeEvent, err
}
