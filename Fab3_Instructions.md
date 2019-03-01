# Fab3 Instruction Set

Fab3 is a partial implementation of the Ethereum JSON RPC API. Requests are
expected in the format and most always have a POST header.

```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": <method>,
  "id":<client-generated-id>,
  "params":<method-params>
}'
```
The examples below use a SimpleStorage contract.
Fab3 currntly supports:
- [net_version](#net_version)
- [eth_getCode](#eth_getCode)
- [eth_call](#eth_call)
- [eth_sendTransaction](#eth_sendTransaction)
- [eth_accounts](#eth_accounts)
- [eth_estimateGas](#eth_estimateGas)
- [eth_getBalance](#eth_getBalance)
- [eth_getBlockByNumber](#eth_getBlockByNumber)
- [eth_blockNumber](#eth_blockNumber)
- [eth_getTransactionByHash](#eth_getTransactionByHash)
- [eth_getTransactionReceipt](#eth_getTransactionReceipt)


### net_version
`net_version` takes no arguments and returns the string `fabric-evm`.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "net_version",
  "id":1,
  "params":[]
}'

{"jsonrpc":"2.0","result":"fabric-evm","id":1}
```


### eth_getCode
`eth_getCode` returns the runtime bytecode of the provided contract address.
According to the spec, getCode takes in two arguments, the first is the contract
address and the second is the block number specifying the state of the ledger to
run the query.

Fab3 does not support querying the state at a certain point in the ledger so the
second argument, if provided, will be ignored. Only the first argument, the
contract address, is required and honored by fab3.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_getCode",
  "id":1,
  "params":["0x40421fd8b64e91da48e703ea1daa488b44ff9d16"]
}'

{"jsonrpc":"2.0","result":"6080604052600436106043576000357c01000000000000000000000000000000000000000000000000000000009004806360fe47b11460485780636d4ce63c14607f575b600080fd5b348015605357600080fd5b50607d60048036036020811015606857600080fd5b810190808035906020019092919050505060a7565b005b348015608a57600080fd5b50609160b1565b6040518082815260200191505060405180910390f35b8060008190555050565b6000805490509056fea165627a7a72305820290b24d16ffaf96310c5e236cef6f8bd81744b72beaeae1ca817d9372b69c2ba0029","id":1}
```

### eth_call
`eth_call` queries the deployed EVMCC and simulates the transaction associated
with the specified parameters. According to the spec, call takes in two
arguments, the first is an object specifying the parameters of the transaction
and the second is the block number specifying the state of the ledger to run
the query against.

Only the first object is required and honored by fab3. The fields `to`, `data`
are the only fields that are required in the object and the rest are ignored if
provided.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_call",
  "id":1,
  "params":[{"to":"0x40421fd8b64e91da48e703ea1daa488b44ff9d16", "data":"0x6d4ce63c"}]
}'

{"jsonrpc":"2.0","result":"0x000000000000000000000000000000000000000000000000000000000000000a","id":1}
```

### eth_sendTransaction
`eth_sendTransaction` submits a transaction to the EVMCC with the specified
parameters. According to the spec, sendTransaction takes an object specifying
the parameters of the transaction. The fields `to`, `data` are the only fields
that are required in the object and the rest are ignored if provided. The fabric
transaction id associated to the transaction is returned

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_sendTransaction",
  "id":1,
  "params":[
    {"to":"0x40421fd8b64e91da48e703ea1daa488b44ff9d16",
    "data":"0x60fe47b1000000000000000000000000000000000000000000000000000000000000000f"}]
}'

{"jsonrpc":"2.0","result":"9807a7ff4ed1962e9414b04f9dec7e05112382a6d826b7e64628fb7f12632dc5","id":1}
```

### eth_accounts
`eth_accounts` queries the EVMCC for the address that is generated from the user
associated to the fab3 instance. The return value will always only have one
address. This method does not take in any parameters. Note the returned address
is generated on the fly by the evmcc and is not stored in the ledger.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_accounts",
  "id":1,
  "params":[]
}'

{"jsonrpc":"2.0","result":["0x564fbd2e6e26ca8dbbac758f9253dd80d90974b6"],"id":1}
```

### eth_estimateGas
Gas is hardcoded in the EVMCC and enough is provided for transactions to
complete. Therefore `eth_estimateGas` will always return 0 regardless of the
parameters that are passed in. Though the parameters are ignored, if provided
they are still expected in the object format otherwise an error will be
returned.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_estimateGas",
  "id":1,
  "params":[
    {"to":"0x40421fd8b64e91da48e703ea1daa488b44ff9d16",
    "data":"0x60fe47b1000000000000000000000000000000000000000000000000000000000000000f"}]
}'

{"jsonrpc":"2.0","result":"0x0","id":1}
```
### eth_getBalance
No ether or native tokens are created as part of the EVMCC. User accounts do not
have any balances. Therefore `eth_getBalance` will always return 0 regardless of
the parameters that are provided. Though the parameters are ignored, an array of
strings is still the expected format. If something else is provided an error
will be returned.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_getBalance",
  "id":1,
  "params":["0x564fbd2e6e26ca8dbbac758f9253dd80d90974b6"]
}'

{"jsonrpc":"2.0","result":"0x0","id":1}
```

### eth_getBlockByNumber
`eth_getBlockByNumber` returns information about the requested block. The method
accepts a number that is hex encoded or a default block parameter such as
`latest`, and `earliest` and a second parameter which is a boolean that
indicates whether full transaction information should be returned. Fabric does
not have a concept of `pending` blocks so providing `pending` as the block
number will result in an error.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_getBlockByNumber",
  "id":1,
  "params":["latest", true]
}'

{
  "jsonrpc": "2.0",
  "result": {
    "number": "0x6",
    "hash": "0x91e7c644378b5386ce317ad7e55f57f230e4234488c88fbb37d170a3d110c55c",
    "parentHash": "0x477617af182bfa77690bc07d0d9e301192474740d2dee4d6c56ba3a6483a11fa",
    "transactions": [
      {
        "blockHash": "0x91e7c644378b5386ce317ad7e55f57f230e4234488c88fbb37d170a3d110c55c",
        "blockNumber": "0x6",
        "to": "0x40421fd8b64e91da48e703ea1daa488b44ff9d16",
        "input": "0x60fe47b1000000000000000000000000000000000000000000000000000000000000000f",
        "transactionIndex": "0x0",
        "hash": "0x9807a7ff4ed1962e9414b04f9dec7e05112382a6d826b7e64628fb7f12632dc5"
      }
    ]
  },
  "id": 1
}
```
### eth_blockNumber
`eth_blockNumber` returns the number associated with the latest block on the
ledger. This method accepts no parameters.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_BlockNumber",
  "id":1,
  "params":[]
}'

{"jsonrpc":"2.0","result":"0x6","id":1}
```

### eth_getTransactionByHash
`eth_getTransactionByHash` will return transaction information about the given
Fabric transaction id. It accepts one parameter, the transaction id.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_getTransactionByHash",
  "id":1,
  "params":["0x9807a7ff4ed1962e9414b04f9dec7e05112382a6d826b7e64628fb7f12632dc5"]
}'

{
  "jsonrpc": "2.0",
  "result": {
    "blockHash": "0x91e7c644378b5386ce317ad7e55f57f230e4234488c88fbb37d170a3d110c55c",
    "blockNumber": "0x6",
    "to": "0x40421fd8b64e91da48e703ea1daa488b44ff9d16",
    "input": "0x60fe47b1000000000000000000000000000000000000000000000000000000000000000f",
    "transactionIndex": "0x0",
    "hash": "0x9807a7ff4ed1962e9414b04f9dec7e05112382a6d826b7e64628fb7f12632dc5"
  },
  "id": 1
}
```

### eth_getTransactionReceipt
`eth_getTransactionReceipt` returns the receipt for the transaction. This
includes any logs that were generated from the transaction. If the transaction
was a contract creation, it will return the contract address of the newly
created contract. Otherwise the contract addresss will be null. This method
accepts only one parameter the Fabric transaction id.

**Example**
```
curl http://127.0.0.1:5000 -X POST -H "Content-Type:application/json" -d '{
  "jsonrpc":"2.0",
  "method": "eth_getTransactionReceipt",
  "id":1,
  "params":["0x39221cdec040293d3124a83e03d8e5555442a5a56ce69a0f866e29fd545f76f5"]
}'

{
  "jsonrpc": "2.0",
  "result": {
    "transactionHash": "0x39221cdec040293d3124a83e03d8e5555442a5a56ce69a0f866e29fd545f76f5",
    "transactionIndex": "0x0",
    "blockHash": "0x0b1cbfa3fa4a5963f025b503ee41cef7500090a4fb102fd569672c17cf2b7f9d",
    "blockNumber": "0x2",
    "contractAddress": "40421fd8b64e91da48e703ea1daa488b44ff9d16",
    "gasUsed": 0,
    "cumulativeGasUsed": 0,
    "to": "",
    "logs": null,
    "status": "0x1"
  },
  "id": 1
}
```
