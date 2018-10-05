/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab3

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/hyperledger/fabric-chaincode-evm/integration/helpers"
	"github.com/hyperledger/fabric/integration/nwo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit"
)

func sendRPCRequest(client *http.Client, method, proxyAddress string, id int, params interface{}) (*http.Response, error) {
	request := helpers.JsonRPCRequest{
		JsonRPC: "2.0",
		Method:  method,
		ID:      id,
		Params:  params,
	}
	reqBody, err := json.Marshal(request)
	Expect(err).ToNot(HaveOccurred())

	body := strings.NewReader(string(reqBody))
	req, err := http.NewRequest("POST", proxyAddress, body)
	Expect(err).ToNot(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

var _ = Describe("Fabproxy Integration", func() {
	var (
		testDir         string
		dockerClient    *docker.Client
		network         *nwo.Network
		chaincode       nwo.Chaincode
		process         ifrit.Process
		channelName     string
		proxyConfigPath string
		ccid            string

		proxy        ifrit.Process
		proxyAddress string

		client        *http.Client
		SimpleStorage helpers.Contract
	)

	BeforeEach(func() {
		var err error
		testDir, err = ioutil.TempDir("", "fab3-e2e")
		Expect(err).NotTo(HaveOccurred())

		dockerClient, err = docker.NewClientFromEnv()
		Expect(err).NotTo(HaveOccurred())

		SimpleStorage = helpers.SimpleStorageContract()

		client = &http.Client{}

		ccid = "evmcc"
		chaincode = nwo.Chaincode{
			Name:    ccid,
			Version: "0.0",
			Path:    "github.com/hyperledger/fabric-chaincode-evm/evmcc",
			Ctor:    `{"Args":[]}`,
			Policy:  `AND ('Org1MSP.member','Org2MSP.member')`,
		}
		network = nwo.New(nwo.BasicSolo(), testDir, dockerClient, 30000, components)
		network.GenerateConfigTree()
		network.Bootstrap()

		networkRunner := network.NetworkGroupRunner()
		process = ifrit.Invoke(networkRunner)
		Eventually(process.Ready()).Should(BeClosed())
		channelName = "testchannel"

		proxyConfigPath, err = helpers.CreateProxyConfig(testDir, channelName, network.CryptoPath(),
			network.PeerPort(network.Peer("Org1", "peer0"), nwo.ListenPort),
			network.PeerPort(network.Peer("Org1", "peer1"), nwo.ListenPort),
			network.PeerPort(network.Peer("Org2", "peer0"), nwo.ListenPort),
			network.PeerPort(network.Peer("Org2", "peer1"), nwo.ListenPort),
			network.OrdererPort(network.Orderer("orderer"), nwo.ListenPort),
		)
		Expect(err).ToNot(HaveOccurred())

		//Set up the network
		By("getting the orderer by name")
		orderer := network.Orderer("orderer")

		By("setting up the channel")
		network.CreateAndJoinChannel(orderer, "testchannel")
		network.UpdateChannelAnchors(orderer, "testchannel")

		By("deploying the chaincode")
		nwo.DeployChaincode(network, "testchannel", orderer, chaincode)

		By("starting up the proxy")
		proxyPort := network.ReservePort()
		proxyRunner := helpers.FabProxyRunner(components.Paths["fabproxy"], proxyConfigPath, "Org1", "User1", channelName, ccid, proxyPort)
		proxy = ifrit.Invoke(proxyRunner)
		Eventually(proxy.Ready(), 15*time.Second).Should(BeClosed())
		proxyAddress = fmt.Sprintf("http://127.0.0.1:%d", proxyPort)
	})

	AfterEach(func() {
		if process != nil {
			process.Signal(syscall.SIGTERM)
			Eventually(process.Wait(), time.Minute).Should(Receive())
		}
		if network != nil {
			network.Cleanup()
		}
		if proxy != nil {
			proxy.Signal(syscall.SIGTERM)
			Eventually(proxy.Wait(), time.Minute).Should(Receive())
		}
		os.RemoveAll(testDir)
	})

	It("implements the ethereum json rpc api", func() {
		By("querying for an account")
		resp, err := sendRPCRequest(client, "eth_accounts", proxyAddress, 5, []interface{}{})

		expectedArrayBody := helpers.JsonRPCArrayResponse{JsonRPC: "2.0", ID: 5}

		rBody, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		var respArrayBody helpers.JsonRPCArrayResponse
		err = json.Unmarshal(rBody, &respArrayBody)
		Expect(err).ToNot(HaveOccurred())

		Expect(respArrayBody.Error).To(Equal(helpers.JsonRPCError{}))

		Expect(respArrayBody.Result).To(HaveLen(1))
		account := respArrayBody.Result[0]

		checkHexEncoded(account)
		// Set the same result so that next expectation can check all other fields
		expectedArrayBody.Result = respArrayBody.Result
		Expect(respArrayBody).To(Equal(expectedArrayBody))

		By("Deploying the Simple Storage Contract")
		params := helpers.MessageParams{
			To:   "0000000000000000000000000000000000000000",
			Data: SimpleStorage.CompiledBytecode,
		}

		resp, err = sendRPCRequest(client, "eth_sendTransaction", proxyAddress, 6, params)
		Expect(err).ToNot(HaveOccurred())

		expectedBody := helpers.JsonRPCResponse{JsonRPC: "2.0", ID: 6}
		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		var respBody helpers.JsonRPCResponse
		err = json.Unmarshal(rBody, &respBody)
		Expect(err).ToNot(HaveOccurred())

		Expect(respBody.Error).To(Equal(helpers.JsonRPCError{}))

		// Set the same result so that next expectation can check all other fields
		expectedBody.Result = respBody.Result
		Expect(respBody).To(Equal(expectedBody))

		By("Getting the Transaction Receipt")
		txHash := respBody.Result

		resp, err = sendRPCRequest(client, "eth_getTransactionReceipt", proxyAddress, 16, []string{txHash})
		Expect(err).ToNot(HaveOccurred())

		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		var rpcResp helpers.JsonRPCTxReceipt
		err = json.Unmarshal(rBody, &rpcResp)
		Expect(err).ToNot(HaveOccurred())

		Expect(rpcResp.Error).To(Equal(helpers.JsonRPCError{}))

		receipt := rpcResp.Result

		Expect(receipt.ContractAddress).ToNot(Equal(""))
		Expect(receipt.TransactionHash).To(Equal(txHash))
		checkHexEncoded(receipt.BlockNumber)
		checkHexEncoded(receipt.BlockHash)
		checkHexEncoded(receipt.TransactionIndex)

		By("verifying the code")
		contractAddr := receipt.ContractAddress
		resp, err = sendRPCRequest(client, "eth_getCode", proxyAddress, 17, []string{contractAddr})
		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(rBody, &respBody)
		Expect(err).ToNot(HaveOccurred())

		Expect(rpcResp.Error).To(Equal(helpers.JsonRPCError{}))
		Expect(respBody.Result).To(Equal(SimpleStorage.RuntimeBytecode))

		By("interacting with the contract")
		val := "000000000000000000000000000000000000000000000000000000000000002a"
		params = helpers.MessageParams{
			To:   contractAddr,
			Data: SimpleStorage.FunctionHashes["set"] + val,
		}
		resp, err = sendRPCRequest(client, "eth_sendTransaction", proxyAddress, 18, params)
		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(rBody, &respBody)
		Expect(err).ToNot(HaveOccurred())

		Expect(respBody.Error).To(Equal(helpers.JsonRPCError{}))
		txHash = respBody.Result

		By("verifying it returned a valid transaction hash")
		resp, err = sendRPCRequest(client, "eth_getTransactionReceipt", proxyAddress, 16, []string{txHash})
		Expect(err).ToNot(HaveOccurred())

		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(rBody, &rpcResp)
		Expect(err).ToNot(HaveOccurred())

		Expect(rpcResp.Error).To(Equal(helpers.JsonRPCError{}))
		receipt = rpcResp.Result
		Expect(receipt.TransactionHash).To(Equal(txHash))
		checkHexEncoded(receipt.BlockNumber)
		checkHexEncoded(receipt.BlockHash)
		checkHexEncoded(receipt.TransactionIndex)

		// Contract Address field should not be returned for transactions that are not contract deployments
		Expect(string(rBody)).ToNot(ContainSubstring("contractAddress"))

		By("querying the contract")
		params = helpers.MessageParams{
			To:   contractAddr,
			Data: SimpleStorage.FunctionHashes["get"],
		}
		resp, err = sendRPCRequest(client, "eth_call", proxyAddress, 19, params)
		rBody, err = ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		err = json.Unmarshal(rBody, &respBody)
		Expect(err).ToNot(HaveOccurred())

		Expect(respBody.Error).To(Equal(helpers.JsonRPCError{}))
		Expect(respBody.Result).To(Equal(val))
	})
})

func checkHexEncoded(value string) {
	// Check to see that the result is a hexadecimal string
	// Ensure that the prefix is 0x
	Expect(value).To(ContainSubstring("0x"))

	// Ensure the string is hex
	_, err := hex.DecodeString(value[2:])
	Expect(err).ToNot(HaveOccurred())
}
