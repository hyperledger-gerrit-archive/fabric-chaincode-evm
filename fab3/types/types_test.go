/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetLogsArgsUnmarshalJSON(t *testing.T) {
	RegisterTestingT(t)
	{
		// single addr
		bytes := []byte(`{"fromBlock":"0x0","toBlock":"0x1","address":"0x3832333733343538313634383230393437383931"}`)
		var target GetLogsArgs
		err := json.Unmarshal(bytes, &target)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(target.FromBlock).Should(Equal("0"))
		Expect(target.ToBlock).Should(Equal("1"))
		Expect(target.Address).Should(Equal(AddressFilter{"3832333733343538313634383230393437383931"}))
	}
	{
		// array of addr
		bytes := []byte(`{"fromBlock":"0x0","toBlock":"0x1","address":["0x3832333733343538313634383230393437383931"]}`)
		var target GetLogsArgs
		err := json.Unmarshal(bytes, &target)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(target.FromBlock).Should(Equal("0"))
		Expect(target.ToBlock).Should(Equal("1"))
		Expect(target.Address).Should(Equal(AddressFilter{"3832333733343538313634383230393437383931"}))
	}
}
