// Code generated by counterfeiter. DO NOT EDIT.
package fab3

import (
	"net/http"
	"sync"

	"github.com/hyperledger/fabric-chaincode-evm/fab3"
	"github.com/hyperledger/fabric-chaincode-evm/fab3/types"
)

type MockEthService struct {
	AccountsStub        func(*http.Request, *string, *[]string) error
	accountsMutex       sync.RWMutex
	accountsArgsForCall []struct {
		arg1 *http.Request
		arg2 *string
		arg3 *[]string
	}
	accountsReturns struct {
		result1 error
	}
	accountsReturnsOnCall map[int]struct {
		result1 error
	}
	BlockNumberStub        func(*http.Request, *interface{}, *string) error
	blockNumberMutex       sync.RWMutex
	blockNumberArgsForCall []struct {
		arg1 *http.Request
		arg2 *interface{}
		arg3 *string
	}
	blockNumberReturns struct {
		result1 error
	}
	blockNumberReturnsOnCall map[int]struct {
		result1 error
	}
	CallStub        func(*http.Request, *types.EthArgs, *string) error
	callMutex       sync.RWMutex
	callArgsForCall []struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}
	callReturns struct {
		result1 error
	}
	callReturnsOnCall map[int]struct {
		result1 error
	}
	EstimateGasStub        func(*http.Request, *types.EthArgs, *string) error
	estimateGasMutex       sync.RWMutex
	estimateGasArgsForCall []struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}
	estimateGasReturns struct {
		result1 error
	}
	estimateGasReturnsOnCall map[int]struct {
		result1 error
	}
	GetBalanceStub        func(*http.Request, *[]string, *string) error
	getBalanceMutex       sync.RWMutex
	getBalanceArgsForCall []struct {
		arg1 *http.Request
		arg2 *[]string
		arg3 *string
	}
	getBalanceReturns struct {
		result1 error
	}
	getBalanceReturnsOnCall map[int]struct {
		result1 error
	}
	GetBlockByNumberStub        func(*http.Request, *[]interface{}, *types.Block) error
	getBlockByNumberMutex       sync.RWMutex
	getBlockByNumberArgsForCall []struct {
		arg1 *http.Request
		arg2 *[]interface{}
		arg3 *types.Block
	}
	getBlockByNumberReturns struct {
		result1 error
	}
	getBlockByNumberReturnsOnCall map[int]struct {
		result1 error
	}
	GetCodeStub        func(*http.Request, *string, *string) error
	getCodeMutex       sync.RWMutex
	getCodeArgsForCall []struct {
		arg1 *http.Request
		arg2 *string
		arg3 *string
	}
	getCodeReturns struct {
		result1 error
	}
	getCodeReturnsOnCall map[int]struct {
		result1 error
	}
	GetLogsStub        func(*http.Request, *types.GetLogsArgs, *[]types.Log) error
	getLogsMutex       sync.RWMutex
	getLogsArgsForCall []struct {
		arg1 *http.Request
		arg2 *types.GetLogsArgs
		arg3 *[]types.Log
	}
	getLogsReturns struct {
		result1 error
	}
	getLogsReturnsOnCall map[int]struct {
		result1 error
	}
	GetTransactionByHashStub        func(*http.Request, *string, *types.Transaction) error
	getTransactionByHashMutex       sync.RWMutex
	getTransactionByHashArgsForCall []struct {
		arg1 *http.Request
		arg2 *string
		arg3 *types.Transaction
	}
	getTransactionByHashReturns struct {
		result1 error
	}
	getTransactionByHashReturnsOnCall map[int]struct {
		result1 error
	}
	GetTransactionReceiptStub        func(*http.Request, *string, *types.TxReceipt) error
	getTransactionReceiptMutex       sync.RWMutex
	getTransactionReceiptArgsForCall []struct {
		arg1 *http.Request
		arg2 *string
		arg3 *types.TxReceipt
	}
	getTransactionReceiptReturns struct {
		result1 error
	}
	getTransactionReceiptReturnsOnCall map[int]struct {
		result1 error
	}
	SendTransactionStub        func(*http.Request, *types.EthArgs, *string) error
	sendTransactionMutex       sync.RWMutex
	sendTransactionArgsForCall []struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}
	sendTransactionReturns struct {
		result1 error
	}
	sendTransactionReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *MockEthService) Accounts(arg1 *http.Request, arg2 *string, arg3 *[]string) error {
	fake.accountsMutex.Lock()
	ret, specificReturn := fake.accountsReturnsOnCall[len(fake.accountsArgsForCall)]
	fake.accountsArgsForCall = append(fake.accountsArgsForCall, struct {
		arg1 *http.Request
		arg2 *string
		arg3 *[]string
	}{arg1, arg2, arg3})
	fake.recordInvocation("Accounts", []interface{}{arg1, arg2, arg3})
	fake.accountsMutex.Unlock()
	if fake.AccountsStub != nil {
		return fake.AccountsStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.accountsReturns
	return fakeReturns.result1
}

func (fake *MockEthService) AccountsCallCount() int {
	fake.accountsMutex.RLock()
	defer fake.accountsMutex.RUnlock()
	return len(fake.accountsArgsForCall)
}

func (fake *MockEthService) AccountsCalls(stub func(*http.Request, *string, *[]string) error) {
	fake.accountsMutex.Lock()
	defer fake.accountsMutex.Unlock()
	fake.AccountsStub = stub
}

func (fake *MockEthService) AccountsArgsForCall(i int) (*http.Request, *string, *[]string) {
	fake.accountsMutex.RLock()
	defer fake.accountsMutex.RUnlock()
	argsForCall := fake.accountsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) AccountsReturns(result1 error) {
	fake.accountsMutex.Lock()
	defer fake.accountsMutex.Unlock()
	fake.AccountsStub = nil
	fake.accountsReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) AccountsReturnsOnCall(i int, result1 error) {
	fake.accountsMutex.Lock()
	defer fake.accountsMutex.Unlock()
	fake.AccountsStub = nil
	if fake.accountsReturnsOnCall == nil {
		fake.accountsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.accountsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) BlockNumber(arg1 *http.Request, arg2 *interface{}, arg3 *string) error {
	fake.blockNumberMutex.Lock()
	ret, specificReturn := fake.blockNumberReturnsOnCall[len(fake.blockNumberArgsForCall)]
	fake.blockNumberArgsForCall = append(fake.blockNumberArgsForCall, struct {
		arg1 *http.Request
		arg2 *interface{}
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("BlockNumber", []interface{}{arg1, arg2, arg3})
	fake.blockNumberMutex.Unlock()
	if fake.BlockNumberStub != nil {
		return fake.BlockNumberStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.blockNumberReturns
	return fakeReturns.result1
}

func (fake *MockEthService) BlockNumberCallCount() int {
	fake.blockNumberMutex.RLock()
	defer fake.blockNumberMutex.RUnlock()
	return len(fake.blockNumberArgsForCall)
}

func (fake *MockEthService) BlockNumberCalls(stub func(*http.Request, *interface{}, *string) error) {
	fake.blockNumberMutex.Lock()
	defer fake.blockNumberMutex.Unlock()
	fake.BlockNumberStub = stub
}

func (fake *MockEthService) BlockNumberArgsForCall(i int) (*http.Request, *interface{}, *string) {
	fake.blockNumberMutex.RLock()
	defer fake.blockNumberMutex.RUnlock()
	argsForCall := fake.blockNumberArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) BlockNumberReturns(result1 error) {
	fake.blockNumberMutex.Lock()
	defer fake.blockNumberMutex.Unlock()
	fake.BlockNumberStub = nil
	fake.blockNumberReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) BlockNumberReturnsOnCall(i int, result1 error) {
	fake.blockNumberMutex.Lock()
	defer fake.blockNumberMutex.Unlock()
	fake.BlockNumberStub = nil
	if fake.blockNumberReturnsOnCall == nil {
		fake.blockNumberReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.blockNumberReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) Call(arg1 *http.Request, arg2 *types.EthArgs, arg3 *string) error {
	fake.callMutex.Lock()
	ret, specificReturn := fake.callReturnsOnCall[len(fake.callArgsForCall)]
	fake.callArgsForCall = append(fake.callArgsForCall, struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("Call", []interface{}{arg1, arg2, arg3})
	fake.callMutex.Unlock()
	if fake.CallStub != nil {
		return fake.CallStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.callReturns
	return fakeReturns.result1
}

func (fake *MockEthService) CallCallCount() int {
	fake.callMutex.RLock()
	defer fake.callMutex.RUnlock()
	return len(fake.callArgsForCall)
}

func (fake *MockEthService) CallCalls(stub func(*http.Request, *types.EthArgs, *string) error) {
	fake.callMutex.Lock()
	defer fake.callMutex.Unlock()
	fake.CallStub = stub
}

func (fake *MockEthService) CallArgsForCall(i int) (*http.Request, *types.EthArgs, *string) {
	fake.callMutex.RLock()
	defer fake.callMutex.RUnlock()
	argsForCall := fake.callArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) CallReturns(result1 error) {
	fake.callMutex.Lock()
	defer fake.callMutex.Unlock()
	fake.CallStub = nil
	fake.callReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) CallReturnsOnCall(i int, result1 error) {
	fake.callMutex.Lock()
	defer fake.callMutex.Unlock()
	fake.CallStub = nil
	if fake.callReturnsOnCall == nil {
		fake.callReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.callReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) EstimateGas(arg1 *http.Request, arg2 *types.EthArgs, arg3 *string) error {
	fake.estimateGasMutex.Lock()
	ret, specificReturn := fake.estimateGasReturnsOnCall[len(fake.estimateGasArgsForCall)]
	fake.estimateGasArgsForCall = append(fake.estimateGasArgsForCall, struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("EstimateGas", []interface{}{arg1, arg2, arg3})
	fake.estimateGasMutex.Unlock()
	if fake.EstimateGasStub != nil {
		return fake.EstimateGasStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.estimateGasReturns
	return fakeReturns.result1
}

func (fake *MockEthService) EstimateGasCallCount() int {
	fake.estimateGasMutex.RLock()
	defer fake.estimateGasMutex.RUnlock()
	return len(fake.estimateGasArgsForCall)
}

func (fake *MockEthService) EstimateGasCalls(stub func(*http.Request, *types.EthArgs, *string) error) {
	fake.estimateGasMutex.Lock()
	defer fake.estimateGasMutex.Unlock()
	fake.EstimateGasStub = stub
}

func (fake *MockEthService) EstimateGasArgsForCall(i int) (*http.Request, *types.EthArgs, *string) {
	fake.estimateGasMutex.RLock()
	defer fake.estimateGasMutex.RUnlock()
	argsForCall := fake.estimateGasArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) EstimateGasReturns(result1 error) {
	fake.estimateGasMutex.Lock()
	defer fake.estimateGasMutex.Unlock()
	fake.EstimateGasStub = nil
	fake.estimateGasReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) EstimateGasReturnsOnCall(i int, result1 error) {
	fake.estimateGasMutex.Lock()
	defer fake.estimateGasMutex.Unlock()
	fake.EstimateGasStub = nil
	if fake.estimateGasReturnsOnCall == nil {
		fake.estimateGasReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.estimateGasReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetBalance(arg1 *http.Request, arg2 *[]string, arg3 *string) error {
	fake.getBalanceMutex.Lock()
	ret, specificReturn := fake.getBalanceReturnsOnCall[len(fake.getBalanceArgsForCall)]
	fake.getBalanceArgsForCall = append(fake.getBalanceArgsForCall, struct {
		arg1 *http.Request
		arg2 *[]string
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetBalance", []interface{}{arg1, arg2, arg3})
	fake.getBalanceMutex.Unlock()
	if fake.GetBalanceStub != nil {
		return fake.GetBalanceStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getBalanceReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetBalanceCallCount() int {
	fake.getBalanceMutex.RLock()
	defer fake.getBalanceMutex.RUnlock()
	return len(fake.getBalanceArgsForCall)
}

func (fake *MockEthService) GetBalanceCalls(stub func(*http.Request, *[]string, *string) error) {
	fake.getBalanceMutex.Lock()
	defer fake.getBalanceMutex.Unlock()
	fake.GetBalanceStub = stub
}

func (fake *MockEthService) GetBalanceArgsForCall(i int) (*http.Request, *[]string, *string) {
	fake.getBalanceMutex.RLock()
	defer fake.getBalanceMutex.RUnlock()
	argsForCall := fake.getBalanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetBalanceReturns(result1 error) {
	fake.getBalanceMutex.Lock()
	defer fake.getBalanceMutex.Unlock()
	fake.GetBalanceStub = nil
	fake.getBalanceReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetBalanceReturnsOnCall(i int, result1 error) {
	fake.getBalanceMutex.Lock()
	defer fake.getBalanceMutex.Unlock()
	fake.GetBalanceStub = nil
	if fake.getBalanceReturnsOnCall == nil {
		fake.getBalanceReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getBalanceReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetBlockByNumber(arg1 *http.Request, arg2 *[]interface{}, arg3 *types.Block) error {
	fake.getBlockByNumberMutex.Lock()
	ret, specificReturn := fake.getBlockByNumberReturnsOnCall[len(fake.getBlockByNumberArgsForCall)]
	fake.getBlockByNumberArgsForCall = append(fake.getBlockByNumberArgsForCall, struct {
		arg1 *http.Request
		arg2 *[]interface{}
		arg3 *types.Block
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetBlockByNumber", []interface{}{arg1, arg2, arg3})
	fake.getBlockByNumberMutex.Unlock()
	if fake.GetBlockByNumberStub != nil {
		return fake.GetBlockByNumberStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getBlockByNumberReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetBlockByNumberCallCount() int {
	fake.getBlockByNumberMutex.RLock()
	defer fake.getBlockByNumberMutex.RUnlock()
	return len(fake.getBlockByNumberArgsForCall)
}

func (fake *MockEthService) GetBlockByNumberCalls(stub func(*http.Request, *[]interface{}, *types.Block) error) {
	fake.getBlockByNumberMutex.Lock()
	defer fake.getBlockByNumberMutex.Unlock()
	fake.GetBlockByNumberStub = stub
}

func (fake *MockEthService) GetBlockByNumberArgsForCall(i int) (*http.Request, *[]interface{}, *types.Block) {
	fake.getBlockByNumberMutex.RLock()
	defer fake.getBlockByNumberMutex.RUnlock()
	argsForCall := fake.getBlockByNumberArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetBlockByNumberReturns(result1 error) {
	fake.getBlockByNumberMutex.Lock()
	defer fake.getBlockByNumberMutex.Unlock()
	fake.GetBlockByNumberStub = nil
	fake.getBlockByNumberReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetBlockByNumberReturnsOnCall(i int, result1 error) {
	fake.getBlockByNumberMutex.Lock()
	defer fake.getBlockByNumberMutex.Unlock()
	fake.GetBlockByNumberStub = nil
	if fake.getBlockByNumberReturnsOnCall == nil {
		fake.getBlockByNumberReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getBlockByNumberReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetCode(arg1 *http.Request, arg2 *string, arg3 *string) error {
	fake.getCodeMutex.Lock()
	ret, specificReturn := fake.getCodeReturnsOnCall[len(fake.getCodeArgsForCall)]
	fake.getCodeArgsForCall = append(fake.getCodeArgsForCall, struct {
		arg1 *http.Request
		arg2 *string
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetCode", []interface{}{arg1, arg2, arg3})
	fake.getCodeMutex.Unlock()
	if fake.GetCodeStub != nil {
		return fake.GetCodeStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getCodeReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetCodeCallCount() int {
	fake.getCodeMutex.RLock()
	defer fake.getCodeMutex.RUnlock()
	return len(fake.getCodeArgsForCall)
}

func (fake *MockEthService) GetCodeCalls(stub func(*http.Request, *string, *string) error) {
	fake.getCodeMutex.Lock()
	defer fake.getCodeMutex.Unlock()
	fake.GetCodeStub = stub
}

func (fake *MockEthService) GetCodeArgsForCall(i int) (*http.Request, *string, *string) {
	fake.getCodeMutex.RLock()
	defer fake.getCodeMutex.RUnlock()
	argsForCall := fake.getCodeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetCodeReturns(result1 error) {
	fake.getCodeMutex.Lock()
	defer fake.getCodeMutex.Unlock()
	fake.GetCodeStub = nil
	fake.getCodeReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetCodeReturnsOnCall(i int, result1 error) {
	fake.getCodeMutex.Lock()
	defer fake.getCodeMutex.Unlock()
	fake.GetCodeStub = nil
	if fake.getCodeReturnsOnCall == nil {
		fake.getCodeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getCodeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetLogs(arg1 *http.Request, arg2 *types.GetLogsArgs, arg3 *[]types.Log) error {
	fake.getLogsMutex.Lock()
	ret, specificReturn := fake.getLogsReturnsOnCall[len(fake.getLogsArgsForCall)]
	fake.getLogsArgsForCall = append(fake.getLogsArgsForCall, struct {
		arg1 *http.Request
		arg2 *types.GetLogsArgs
		arg3 *[]types.Log
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetLogs", []interface{}{arg1, arg2, arg3})
	fake.getLogsMutex.Unlock()
	if fake.GetLogsStub != nil {
		return fake.GetLogsStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getLogsReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetLogsCallCount() int {
	fake.getLogsMutex.RLock()
	defer fake.getLogsMutex.RUnlock()
	return len(fake.getLogsArgsForCall)
}

func (fake *MockEthService) GetLogsCalls(stub func(*http.Request, *types.GetLogsArgs, *[]types.Log) error) {
	fake.getLogsMutex.Lock()
	defer fake.getLogsMutex.Unlock()
	fake.GetLogsStub = stub
}

func (fake *MockEthService) GetLogsArgsForCall(i int) (*http.Request, *types.GetLogsArgs, *[]types.Log) {
	fake.getLogsMutex.RLock()
	defer fake.getLogsMutex.RUnlock()
	argsForCall := fake.getLogsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetLogsReturns(result1 error) {
	fake.getLogsMutex.Lock()
	defer fake.getLogsMutex.Unlock()
	fake.GetLogsStub = nil
	fake.getLogsReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetLogsReturnsOnCall(i int, result1 error) {
	fake.getLogsMutex.Lock()
	defer fake.getLogsMutex.Unlock()
	fake.GetLogsStub = nil
	if fake.getLogsReturnsOnCall == nil {
		fake.getLogsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getLogsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetTransactionByHash(arg1 *http.Request, arg2 *string, arg3 *types.Transaction) error {
	fake.getTransactionByHashMutex.Lock()
	ret, specificReturn := fake.getTransactionByHashReturnsOnCall[len(fake.getTransactionByHashArgsForCall)]
	fake.getTransactionByHashArgsForCall = append(fake.getTransactionByHashArgsForCall, struct {
		arg1 *http.Request
		arg2 *string
		arg3 *types.Transaction
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetTransactionByHash", []interface{}{arg1, arg2, arg3})
	fake.getTransactionByHashMutex.Unlock()
	if fake.GetTransactionByHashStub != nil {
		return fake.GetTransactionByHashStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getTransactionByHashReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetTransactionByHashCallCount() int {
	fake.getTransactionByHashMutex.RLock()
	defer fake.getTransactionByHashMutex.RUnlock()
	return len(fake.getTransactionByHashArgsForCall)
}

func (fake *MockEthService) GetTransactionByHashCalls(stub func(*http.Request, *string, *types.Transaction) error) {
	fake.getTransactionByHashMutex.Lock()
	defer fake.getTransactionByHashMutex.Unlock()
	fake.GetTransactionByHashStub = stub
}

func (fake *MockEthService) GetTransactionByHashArgsForCall(i int) (*http.Request, *string, *types.Transaction) {
	fake.getTransactionByHashMutex.RLock()
	defer fake.getTransactionByHashMutex.RUnlock()
	argsForCall := fake.getTransactionByHashArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetTransactionByHashReturns(result1 error) {
	fake.getTransactionByHashMutex.Lock()
	defer fake.getTransactionByHashMutex.Unlock()
	fake.GetTransactionByHashStub = nil
	fake.getTransactionByHashReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetTransactionByHashReturnsOnCall(i int, result1 error) {
	fake.getTransactionByHashMutex.Lock()
	defer fake.getTransactionByHashMutex.Unlock()
	fake.GetTransactionByHashStub = nil
	if fake.getTransactionByHashReturnsOnCall == nil {
		fake.getTransactionByHashReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getTransactionByHashReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetTransactionReceipt(arg1 *http.Request, arg2 *string, arg3 *types.TxReceipt) error {
	fake.getTransactionReceiptMutex.Lock()
	ret, specificReturn := fake.getTransactionReceiptReturnsOnCall[len(fake.getTransactionReceiptArgsForCall)]
	fake.getTransactionReceiptArgsForCall = append(fake.getTransactionReceiptArgsForCall, struct {
		arg1 *http.Request
		arg2 *string
		arg3 *types.TxReceipt
	}{arg1, arg2, arg3})
	fake.recordInvocation("GetTransactionReceipt", []interface{}{arg1, arg2, arg3})
	fake.getTransactionReceiptMutex.Unlock()
	if fake.GetTransactionReceiptStub != nil {
		return fake.GetTransactionReceiptStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getTransactionReceiptReturns
	return fakeReturns.result1
}

func (fake *MockEthService) GetTransactionReceiptCallCount() int {
	fake.getTransactionReceiptMutex.RLock()
	defer fake.getTransactionReceiptMutex.RUnlock()
	return len(fake.getTransactionReceiptArgsForCall)
}

func (fake *MockEthService) GetTransactionReceiptCalls(stub func(*http.Request, *string, *types.TxReceipt) error) {
	fake.getTransactionReceiptMutex.Lock()
	defer fake.getTransactionReceiptMutex.Unlock()
	fake.GetTransactionReceiptStub = stub
}

func (fake *MockEthService) GetTransactionReceiptArgsForCall(i int) (*http.Request, *string, *types.TxReceipt) {
	fake.getTransactionReceiptMutex.RLock()
	defer fake.getTransactionReceiptMutex.RUnlock()
	argsForCall := fake.getTransactionReceiptArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) GetTransactionReceiptReturns(result1 error) {
	fake.getTransactionReceiptMutex.Lock()
	defer fake.getTransactionReceiptMutex.Unlock()
	fake.GetTransactionReceiptStub = nil
	fake.getTransactionReceiptReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) GetTransactionReceiptReturnsOnCall(i int, result1 error) {
	fake.getTransactionReceiptMutex.Lock()
	defer fake.getTransactionReceiptMutex.Unlock()
	fake.GetTransactionReceiptStub = nil
	if fake.getTransactionReceiptReturnsOnCall == nil {
		fake.getTransactionReceiptReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.getTransactionReceiptReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) SendTransaction(arg1 *http.Request, arg2 *types.EthArgs, arg3 *string) error {
	fake.sendTransactionMutex.Lock()
	ret, specificReturn := fake.sendTransactionReturnsOnCall[len(fake.sendTransactionArgsForCall)]
	fake.sendTransactionArgsForCall = append(fake.sendTransactionArgsForCall, struct {
		arg1 *http.Request
		arg2 *types.EthArgs
		arg3 *string
	}{arg1, arg2, arg3})
	fake.recordInvocation("SendTransaction", []interface{}{arg1, arg2, arg3})
	fake.sendTransactionMutex.Unlock()
	if fake.SendTransactionStub != nil {
		return fake.SendTransactionStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.sendTransactionReturns
	return fakeReturns.result1
}

func (fake *MockEthService) SendTransactionCallCount() int {
	fake.sendTransactionMutex.RLock()
	defer fake.sendTransactionMutex.RUnlock()
	return len(fake.sendTransactionArgsForCall)
}

func (fake *MockEthService) SendTransactionCalls(stub func(*http.Request, *types.EthArgs, *string) error) {
	fake.sendTransactionMutex.Lock()
	defer fake.sendTransactionMutex.Unlock()
	fake.SendTransactionStub = stub
}

func (fake *MockEthService) SendTransactionArgsForCall(i int) (*http.Request, *types.EthArgs, *string) {
	fake.sendTransactionMutex.RLock()
	defer fake.sendTransactionMutex.RUnlock()
	argsForCall := fake.sendTransactionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *MockEthService) SendTransactionReturns(result1 error) {
	fake.sendTransactionMutex.Lock()
	defer fake.sendTransactionMutex.Unlock()
	fake.SendTransactionStub = nil
	fake.sendTransactionReturns = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) SendTransactionReturnsOnCall(i int, result1 error) {
	fake.sendTransactionMutex.Lock()
	defer fake.sendTransactionMutex.Unlock()
	fake.SendTransactionStub = nil
	if fake.sendTransactionReturnsOnCall == nil {
		fake.sendTransactionReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.sendTransactionReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MockEthService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.accountsMutex.RLock()
	defer fake.accountsMutex.RUnlock()
	fake.blockNumberMutex.RLock()
	defer fake.blockNumberMutex.RUnlock()
	fake.callMutex.RLock()
	defer fake.callMutex.RUnlock()
	fake.estimateGasMutex.RLock()
	defer fake.estimateGasMutex.RUnlock()
	fake.getBalanceMutex.RLock()
	defer fake.getBalanceMutex.RUnlock()
	fake.getBlockByNumberMutex.RLock()
	defer fake.getBlockByNumberMutex.RUnlock()
	fake.getCodeMutex.RLock()
	defer fake.getCodeMutex.RUnlock()
	fake.getLogsMutex.RLock()
	defer fake.getLogsMutex.RUnlock()
	fake.getTransactionByHashMutex.RLock()
	defer fake.getTransactionByHashMutex.RUnlock()
	fake.getTransactionReceiptMutex.RLock()
	defer fake.getTransactionReceiptMutex.RUnlock()
	fake.sendTransactionMutex.RLock()
	defer fake.sendTransactionMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *MockEthService) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ fab3.EthService = new(MockEthService)
