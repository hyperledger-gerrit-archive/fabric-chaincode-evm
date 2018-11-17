/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package helpers

type Contract struct {
	CompiledBytecode string
	RuntimeBytecode  string
	FunctionHashes   map[string]string
}

func SimpleStorageContract() Contract {
	/* SimpleStorage Contract
	pragma solidity ^0.4.0;

	contract SimpleStorage {
	    uint storedData;

	    function set(uint x) public {
	        storedData = x;
	    }

	    function get() public constant returns (uint) {
	        return storedData;
	    }
	}
	*/
	functionHashes := make(map[string]string)
	functionHashes["get"] = "6d4ce63c"
	functionHashes["set"] = "60fe47b1"

	return Contract{
		CompiledBytecode: "608060405234801561001057600080fd5b5060df8061001f6000396000f3006080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806360fe47b114604e5780636d4ce63c146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820b0d276c882ecf7f9b68cdde8700e1aec4986a0d9db3cd43fd61d3dd67278d6cf0029",
		RuntimeBytecode:  "6080604052600436106049576000357c0100000000000000000000000000000000000000000000000000000000900463ffffffff16806360fe47b114604e5780636d4ce63c146078575b600080fd5b348015605957600080fd5b5060766004803603810190808035906020019092919050505060a0565b005b348015608357600080fd5b50608a60aa565b6040518082815260200191505060405180910390f35b8060008190555050565b600080549050905600a165627a7a72305820b0d276c882ecf7f9b68cdde8700e1aec4986a0d9db3cd43fd61d3dd67278d6cf0029",
		FunctionHashes:   functionHashes,
	}
}

func InvokeContract() Contract {
	/* Invokes a previously deployed SimpleStorage Contract
		pragma solidity ^0.4.18;

	  interface StorageInterface{
	    function get() external returns (uint);
	    function set(uint _val);
	  }

	  contract Invoke{
	    StorageInterface store;

	    constructor(StorageInterface _store) public {
	       store = _store;
	    }

	    function getVal() public view returns (uint result) {
	        return store.get();
	    }

	    function setA(uint _val) public returns (uint result) {
	        store.set(_val);
	        return _val;
	    }
	  }
	*/

	functionHashes := make(map[string]string)
	functionHashes["getVal"] = "e1cb0e52"
	functionHashes["setVal"] = "3d4197f0"

	return Contract{
		CompiledBytecode: "608060405234801561001057600080fd5b50604051602080610256833981016040525160008054600160a060020a03909216600160a060020a0319909216919091179055610204806100526000396000f30060806040526004361061004b5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416633d4197f08114610050578063e1cb0e521461007a575b600080fd5b34801561005c57600080fd5b5061006860043561008f565b60408051918252519081900360200190f35b34801561008657600080fd5b50610068610120565b60008054604080517f60fe47b100000000000000000000000000000000000000000000000000000000815260048101859052905173ffffffffffffffffffffffffffffffffffffffff909216916360fe47b191602480820192869290919082900301818387803b15801561010257600080fd5b505af1158015610116573d6000803e3d6000fd5b5093949350505050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16636d4ce63c6040518163ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401602060405180830381600087803b1580156101a757600080fd5b505af11580156101bb573d6000803e3d6000fd5b505050506040513d60208110156101d157600080fd5b50519050905600a165627a7a723058208873497fb521d304cc79b588e9adad54377e155c77b1e5b3c35f2563657fb0700029",
		RuntimeBytecode:  "60806040526004361061004b5763ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416633d4197f08114610050578063e1cb0e521461007a575b600080fd5b34801561005c57600080fd5b5061006860043561008f565b60408051918252519081900360200190f35b34801561008657600080fd5b50610068610120565b60008054604080517f60fe47b100000000000000000000000000000000000000000000000000000000815260048101859052905173ffffffffffffffffffffffffffffffffffffffff909216916360fe47b191602480820192869290919082900301818387803b15801561010257600080fd5b505af1158015610116573d6000803e3d6000fd5b5093949350505050565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16636d4ce63c6040518163ffffffff167c0100000000000000000000000000000000000000000000000000000000028152600401602060405180830381600087803b1580156101a757600080fd5b505af11580156101bb573d6000803e3d6000fd5b505050506040513d60208110156101d157600080fd5b50519050905600a165627a7a723058208873497fb521d304cc79b588e9adad54377e155c77b1e5b3c35f2563657fb0700029",
		FunctionHashes:   functionHashes,
	}
}
