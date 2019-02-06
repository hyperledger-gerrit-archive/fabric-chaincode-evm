/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Package types contains the types used to interact with the json-rpc
interface. It exists for users of fab3 types to use them without importing the
fabric protobuf definitions.
*/
package types

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

/*
Input types used as arguments to ethservice methods.
*/

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
	FromBlock string        `json:"fromBlock,omitempty"`
	ToBlock   string        `json:"toBlock,omitempty"`
	Address   AddressFilter `json:"address,omitempty"`
	Topics    TopicsFilter  `json:"topics,omitempty"`
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

	if input.Address != nil {
		var af AddressFilter
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
	}

	if input.Topics != nil {
		var tf TopicsFilter

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
					var mtf TopicFilter
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
	}

	return nil
}

const (
	HexEncodedAddressLegnth = 42 // 20 bytes, is 40 hex chars, plus two for '0x'
	HexEncodedTopicLegnth   = 66 // 32 bytes, is 64 hex chars, plus two for '0x'
)

type AddressFilter []string // 20 Byte Addresses

// NewAddressFilter takes a string and checks that is the correct length to
// represent a topic and strips the 0x
func NewAddressFilter(s string) (AddressFilter, error) {
	if len(s) != HexEncodedAddressLegnth {
		return nil, fmt.Errorf("address in wrong format, not 42 chars %q", s)
	}
	//Not checking for malformed addresses just stripping `0x` prefix where applicable
	if len(s) > 2 && s[0:2] == "0x" {
		s = s[2:]
	}
	return AddressFilter{s}, nil
}

type TopicFilter []string // 32 Byte Topics

// NewTopicFilter takes a string and checks that is the correct length to
// represent a topic and strips the 0x
func NewTopicFilter(s string) (TopicFilter, error) {
	if len(s) != HexEncodedTopicLegnth {
		return nil, fmt.Errorf("topic in wrong format, %d chars not 66 chars %q", len(s), s)
	}
	//Not checking for malformed addresses just stripping `0x` prefix where applicable
	if len(s) > 2 && s[0:2] == "0x" {
		s = s[2:]
	}
	return TopicFilter{s}, nil
}

// slices are in order, but maybe we should name individual fields first,second,
// etc.  only a few individual fields.
//
// TODO(mhb): maybe an alias for the contained type instead of a struct?
type TopicsFilter []TopicFilter

func NewTopicsFilter(tf ...TopicFilter) TopicsFilter {
	return tf
}

/*
Output types used as return values from ethservice methods.
*/

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
