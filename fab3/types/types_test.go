/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/extensions/table"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// slack input validation refers to having a 0x as prefix for input values

var _ = Describe("GetLogsArgs UnmarshalJSON", func() {

	type valid struct {
		from      string
		to        string
		addresses []string
		topics    [][]string
	}
	DescribeTable("Valid JSON",
		func(bytes []byte, check valid) {
			var target GetLogsArgs
			err := json.Unmarshal(bytes, &target)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(target.FromBlock).Should(Equal(check.from))
			Expect(target.ToBlock).Should(Equal(check.to))
			for _, address := range check.addresses {
				Expect(target.Address).Should(ContainElement(address)) // order not important
				fmt.Fprintln(GinkgoWriter, target.Address, address)
			}
			fmt.Fprintln(GinkgoWriter, target.Topics, check.topics)
			for i, topicFilters := range check.topics {
				for _, topic := range topicFilters {
					fmt.Fprintln(GinkgoWriter, target.Topics[i], topic)
					Expect(target.Topics[i]).Should(ContainElement(topic))
				}
			}
		},

		Entry("empty json",
			[]byte(`{}`),
			valid{"", "", nil, nil}),
		// block references
		Entry("block refs",
			[]byte(`{"fromBlock":"0x0","toBlock":"0x1"}`),
			valid{"0", "1", nil, nil}),
		Entry("slack block refs",
			[]byte(`{"fromBlock":"0","toBlock":"1"}`),
			valid{"0", "1", nil, nil}),
		Entry("textual block refs",
			[]byte(`{"fromBlock":"earliest","toBlock":"latest"}`),
			valid{"earliest", "latest", nil, nil}),
		Entry("from-to order not checked, from is greater than to",
			[]byte(`{"fromBlock":"0x1","toBlock":"0x0"}`),
			valid{"1", "0", nil, nil}),
		Entry("garbage block refs not validated",
			[]byte(`{"toBlock":"there","fromBlock":"hello"}`),
			valid{"hello", "there", nil, nil}),
		// addresses
		Entry("single address",
			[]byte(`{"address":"0x3832333733343538313634383230393437383931"}`),
			valid{"", "", []string{"3832333733343538313634383230393437383931"}, nil}),
		Entry("single address in array",
			[]byte(`{"address":["0x3832333733343538313634383230393437383931"]}`),
			valid{"", "", []string{"3832333733343538313634383230393437383931"}, nil}),
		Entry("multiple address in array",
			[]byte(`{"address":["0x3832333733343538313634383230393437383931","0x3832333733343538313634383230393437383932"]}`),
			valid{"", "", []string{"3832333733343538313634383230393437383931", "3832333733343538313634383230393437383932"}, nil}),
		Entry("slack validation of multiple address in array",
			[]byte(`{"address":["3832333733343538313634383230393437383931","3832333733343538313634383230393437383932"]}`),
			valid{"", "", []string{"3832333733343538313634383230393437383931", "3832333733343538313634383230393437383932"}, nil}),
		// topics
		Entry("any topic",
			[]byte(`{"topics":[]}`),
			valid{"", "", nil, [][]string{{}}}),
		Entry("single topic",
			[]byte(`{"topics":["0x1234567890123456789012345678901234567890123456789012345678901234"]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}}}),
		Entry("single topic, slack input",
			[]byte(`{"topics":["1234567890123456789012345678901234567890123456789012345678901234"]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}}}),
		Entry("single topic with or'd options",
			[]byte(`{"topics":[["0x1234567890123456789012345678901234567890123456789012345678901234","0x1234567890123456789012345678901234567890123456789012345678901235"]]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234", "1234567890123456789012345678901234567890123456789012345678901235"}}}),
		Entry("single topic with or'd options, slack input",
			[]byte(`{"topics":[["1234567890123456789012345678901234567890123456789012345678901234","1234567890123456789012345678901234567890123456789012345678901235"]]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234", "1234567890123456789012345678901234567890123456789012345678901235"}}}),

		Entry("single topic in multi array",
			[]byte(`{"topics":[["0x1234567890123456789012345678901234567890123456789012345678901234"]]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}}}),
		Entry("single topic in multi array, slack input",
			[]byte(`{"topics":[["1234567890123456789012345678901234567890123456789012345678901234"]]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}}}),
		Entry("multi topic",
			[]byte(`{"topics":["0x1234567890123456789012345678901234567890123456789012345678901234", "0x1234567890123456789012345678901234567890123456789012345678901234"]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}, {"1234567890123456789012345678901234567890123456789012345678901234"}}}),
		Entry("multi topic, slack input",
			[]byte(`{"topics":["1234567890123456789012345678901234567890123456789012345678901234", "1234567890123456789012345678901234567890123456789012345678901234"]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234"}, {"1234567890123456789012345678901234567890123456789012345678901234"}}}),

		Entry("multi topic with or'd options, some slack mixed in",
			[]byte(`{"topics":[["0x1234567890123456789012345678901234567890123456789012345678901234","1234567890123456789012345678901234567890123456789012345678901235"], ["1234567890123456789012345678901234567890123456789012345678901234","0x1234567890123456789012345678901234567890123456789012345678901235"]]}`),
			valid{"", "", nil, [][]string{{"1234567890123456789012345678901234567890123456789012345678901234", "1234567890123456789012345678901234567890123456789012345678901235"}, {"1234567890123456789012345678901234567890123456789012345678901234", "1234567890123456789012345678901234567890123456789012345678901235"}}}),
		// everything
		Entry("all fields, no matter the order",
			[]byte(`{"address":["0x3832333733343538313634383230393437383931","0x3832333733343538313634383230393437383932"],"fromBlock":"0x1","toBlock":"earliest"}`),
			valid{"1", "earliest", []string{"3832333733343538313634383230393437383931", "3832333733343538313634383230393437383931"}, nil}),
	)

	DescribeTable("Invalid input JSON",
		func(bytes []byte) {
			var target GetLogsArgs
			err := json.Unmarshal(bytes, &target)
			Expect(err).Should(HaveOccurred())
		},
		Entry("wrong type",
			[]byte(`{"fromBlock":4}`)),
		Entry("not json",
			[]byte(`adrsen@(*P@*#J`)),
		Entry("bad formatted single address",
			[]byte(`{"address":"343538313634383230393437383932"}`)),
		Entry("bad formatted addresses",
			[]byte(`{"address":["0x38323","343538313634383230393437383932"]}`)),
		Entry("mixed good and bad addresses",
			[]byte(`{"address":["0x3832333733343538313634383230393437383931","634383230393437383932"]}`)),
		Entry("empty single address",
			[]byte(`{"address":""}`)),
		Entry("empty address in array",
			[]byte(`{"address":["0x3832333733343538313634383230393437383931",""]}`)),
		Entry("bad formatted addresses",
			[]byte(`{"address":1233456}`)),
		Entry("must be slice",
			[]byte(`{"topics":"must be slice"}`)),
		Entry("bad single topic",
			[]byte(`{"topics":["not a topic"]}`)),
		Entry("bad array topic",
			[]byte(`{"topics":[["not a topic"]]}`)),
		Entry("non string array topic",
			[]byte(`{"topics":[[1234]]}`)),
		Entry("unparsable trash object",
			[]byte(`{"topics":[{"trash":"object"}]}`)),
	)
})
