# Hyperledger Fabric EVM chaincode

This is the project for the Hyperledger Fabric chaincode, integrating the
Burrow EVM. At its essence, this project enables one to use the Hyperledger
Fabric permissioned blockchain platform to interact with Ethereum smart
contracts written in an EVM compatible language such as Solidity or Vyper.

The integration has two main pieces. The chaincode, which integrates the
Hyperledger Burrow EVM package in a Go chaincode shim and maps the various
methods between the peer and the EVM itself.

The second piece is Fab3, a web3 provider, which is a proxy that implements a
subset of the Ethereum compliant JSON RPC interfaces, so that users could use
tools such as Web3.js to interact with smart contracts running in the Fabric
EVM.

We hang out in the
[#fabric-evm channel](https://chat.hyperledger.org/channel/fabric-evm). We are
always interested in feedback and help in development and testing! For more
information about contributing look below at the [Contributions](#Contributions)
section.


## Deploying the Fabric EVM Chaincode (EVMCC)

This chaincode can be deployed like any other user chaincode to Hyperledger
Fabric. The chaincode has no instantiation arguments.

When installing, point to the evmcc [main package](https://github.com/hyperledger/fabric-chaincode-evm/tree/master/evmcc). Below is an example of installation and
instantiation through the peer cli.
```
 peer chaincode install -n evmcc -l golang -v 0 -p github.com/hyperledger/fabric-chaincode-evm/evmcc
 peer chaincode instantiate -n evmcc -v 0 -C <channel-name> -c '{"Args":[]}' -o <orderer-address> --tls --cafile <orderer-ca>
```

The interaction is the same as with any other chaincode, except that
the first argument of a chaincode invoke is the address for the contract and
the second argument is the input you typically provide for an Ethereum
transaction. In general the inputs match the `To` and `Data` fields in an
ethereum transaction. Typically the `To` field is the contract address that
contains the desired transaction. The `Data` field is the function to be invoked
concatenated with the encoded parameters expected.
```
peer chaincode invoke -n evmcc -C <channel-name> -c '{"Args":[<to>,<data>]}' -o <orderer-address> --tls --cafile <orderer-ca>

# Contract Invocation
peer chaincode invoke -n evmcc -C <channel-name> -c '{"Args":[<contract-address>,<function-with-encoded-params>]}' -o <orderer-address> --tls --cafile <orderer-ca>
```
A special case of the ethereum transaction is contract creation. The `To` field
is the zero address and the `Data` field is the compiled bytecode of the smart
contract to be deployed.
```
# Contract Creation
peer chaincode invoke -n evmcc -C <channel-name> -c '{"Args":["0000000000000000000000000000000000000000",<compiled-bytecode>]}' -o <orderer-address> --tls --cafile <orderer-ca>
```

The only actions that do not follow the above pattern are to query for contract
runtime code and accounts.
```
# To query for the user account address that is generated from the user public key
peer chaincode query -n evmcc -C <channel-name> -c '{"Args":["account"]'

# To query for the runtime bytecode for a contract
peer chaincode query -n evmcc -C <channel-name> -c '{"Args":["getCode", "<contract-address>"]}'
```

**NOTE** No ether or token balance is associated with user accounts, so Ethereum
smart contracts that require a native token cannot be migrated to Fabric
and must be rewritten. Token contracts such as those that follow the ERC 20 standard
can still be deployed to the evmcc.

## Running Fab3

Fab3 is a web3 provider that allows the use of ethereum tools such as Web3.js
and the Remix IDE to interact with the Ethereum Smart Contracts that have been
deployed through the Fabric EVM Chaincode.

For the most up to date instruction set look at the [ethservice](fab3/ethservice.go)
and [netservice](fab3/netservice.go)
implementations. For details about limitations and implementations of the the
instructions look [here](Fab3_Instructions.md).

To create the Fab3 binary, run the following at the root of this repository:
```
make fab3
```
A binary name `fab3` will be located in the `bin` directory.

To run Fab3, user need to provide the a Fabric SDK Config, the organization and
user the Fab3 should use from the credentials provided in the SDK config.
An example of a config that can be used with the first network example from the
[fabric-samples](https://github.com/hyperledger/fabric-samples) repository can
be found [here](https://github.com/hyperledger/fabric-chaincode-evm/blob/master/examples/first-network-sdk-config.yaml).
The credentials are expected to be in the directory format that the
[cryptogen](https://hyperledger-fabric.readthedocs.io/en/release-1.4/commands/cryptogen.html) binary outputs.

The expected inputs can be provided to fab3 as environment variables or flags.

```
Usage:
  fab3 [flags]

Flags:
  -i, --ccid string      ID of the EVM Chaincode deployed in your fabric network.
                         Can also set FAB3_CCID instead (default "evmcc")

  -C, --channel string   Channel to be used for the transactions.
                         Required if FAB3_CHANNEL is not set

  -c, --config string    Path to a compatible Fabric SDK Go config file.
                         Required if FAB3_CONFIG is not set

  -h, --help             help for fab3

  -o, --org string       Organization of the specified user. Required if FAB3_ORG is not set

  -p, --port int         Port that Fab3 will be running on. Can also set FAB3_PORT (default 5000)

  -u, --user string      User identity being used for the proxy (Matches the users' names in
                         the crypto-config directory specified in the config).
                         Required if FAB3_USER is not set
```

## Tutorial

We have a [tutorial](examples/EVM_Smart_Contracts.md) that runs through the
basic setup of the EVM chaincode as well as setting up fab3. It will also cover
deploying a Solidity contract and interacting with it using the Web3.js node library.

## Testing

You can run the integration tests in which a sample Fabric Network is run and the
chaincode is installed with the CCID: `evmcc`.
```
make integration-test
```
The [end-2-end](integration/e2e/e2e_test.go)
test is derivative of the hyperledger/fabric/integration/e2e test. You can
compare them to see what is different.

The [fab3](integration/fab3/fab3_test.go)
test focuses on the JSON RPC API compatibility. The [web3](integration/fab3/web3_e2e_test.js)
test uses the Web3 node.js library as a client to run tests against fab3 and the evmcc.

## Contributions
The `fabric-chaincode-evm` lives in a [gerrit repository](https://gerrit.hyperledger.org/r/#/admin/projects/fabric-chaincode-evm).
The github repository is a mirror. For more information on how to contribute
look at [Fabric's CONTRIBUTING documentation](http://hyperledger-fabric.readthedocs.io/en/latest/CONTRIBUTING.html).

Please send all pull requests to the gerrit repository. For issues, open a ticket in
the Hyperledger Fabric [JIRA](https://jira.hyperledger.org/projects/FAB/issues)
and add `fabric-chaincode-evm` in the component field.

Current Dependencies:
- Hyperledger Fabric [v1.4](https://github.com/hyperledger/fabric/releases/tag/v1.4.0)
- Hyperledger Fabric SDK Go [revision = "beccd9cb1450fddfe426616e151d709c99f7ccdd"](https://github.com/hyperledger/fabric-sdk-go/tree/beccd9cb1450fddfe426616e151d709c99f7ccdd)
- Dep [v0.5](https://github.com/golang/dep/releases/tag/v0.5.0)
- Go 1.10

[![Creative Commons License](https://i.creativecommons.org/l/by/4.0/88x31.png)](http://creativecommons.org/licenses/by/4.0/)<br>
This work is licensed under a [Creative Commons Attribution 4.0 International License](http://creativecommons.org/licenses/by/4.0/)
