/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabproxy_test

import (
	"net/http"

	"github.com/hyperledger/fabric-chaincode-evm/fabproxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NetService", func() {
	var (
		netservice fabproxy.NetService
	)

	BeforeEach(func() {
		netservice = fabproxy.NetService{}
	})

	Describe("implements Version", func() {
		It("by returning fabric evm network id", func() {
			var reply string
			err := netservice.Version(&http.Request{}, nil, &reply)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(reply).Should(Equal(fabproxy.NetworkID))
		})
	})
})
