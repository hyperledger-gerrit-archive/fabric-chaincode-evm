/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab3_configuration_test

import (
	"fmt"
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Fab3 Configuration", func() {
	var (
		proxyCmd  *exec.Cmd
		proxyPort uint16
	)

	BeforeEach(func() {
		proxyPort = uint16(6000 + config.GinkgoConfig.ParallelNode)
	})

	AfterEach(func() {
		if proxyCmd != nil && proxyCmd.Process != nil {
			proxyCmd.Process.Kill()
		}
	})

	It("can be configured with environment variables", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"])
		proxyCmd.Env = append(proxyCmd.Env, fmt.Sprintf("FAB3_CONFIG=%s", components.Paths["Fab3Config"]))
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_ORG=Org1")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_USER=User1")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_CHANNEL=testchannel")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_CCID=evmcc")
		proxyCmd.Env = append(proxyCmd.Env, fmt.Sprintf("PORT=%d", proxyPort))

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(output).Should(gbytes.Say("Starting Fab3 on port"))
	})

	It("can be configured with flags", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--org", "Org1",
			"--user", "User1",
			"--channel", "testchannel",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(output).Should(gbytes.Say("Starting Fab3 on port"))
	})

	It("will use flag values over environment variables ", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--org", "Org1",
			"--user", "User1",
			"--channel", "testchannel",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		proxyCmd.Env = append(proxyCmd.Env, "FAB3_CONFIG=non-existent-config-path")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_ORG=non-existent-org")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_USER=non-existent-user")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_CHANNEL=non-existent-channel")
		proxyCmd.Env = append(proxyCmd.Env, "FAB3_CCID=non-existent-ccid")
		proxyCmd.Env = append(proxyCmd.Env, "PORT=fake-port")

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(output).Should(gbytes.Say("Starting Fab3 on port"))
	})

	It("requires config to be set", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--org", "Org1",
			"--user", "User1",
			"--channel", "testchannel",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		exitErr := proxyCmd.Wait()
		Expect(exitErr).To(HaveOccurred())
		Eventually(output).Should(gbytes.Say("Missing config"))
	})

	It("requires org to be set", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--user", "User1",
			"--channel", "testchannel",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		exitErr := proxyCmd.Wait()
		Expect(exitErr).To(HaveOccurred())
		Eventually(output).Should(gbytes.Say("Missing org"))
	})

	It("requires user to be set", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--org", "Org1",
			"--channel", "testchannel",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		exitErr := proxyCmd.Wait()
		Expect(exitErr).To(HaveOccurred())
		Eventually(output).Should(gbytes.Say("Missing user"))
	})

	It("requires channel to be set", func() {
		proxyPort := network.ReservePort()

		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--org", "Org1",
			"--user", "User1",
			"--ccid", "evmcc",
			"--port", strconv.FormatUint(uint64(proxyPort), 10),
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		exitErr := proxyCmd.Wait()
		Expect(exitErr).To(HaveOccurred())
		Eventually(output).Should(gbytes.Say("Missing channel"))
	})

	It("does not require ccid or port", func() {
		proxyCmd = exec.Command(components.Paths["fab3"],
			"--config", components.Paths["Fab3Config"],
			"--org", "Org1",
			"--user", "User1",
			"--channel", "testchannel",
		)

		output := gbytes.NewBuffer()
		proxyCmd.Stdout = output
		proxyCmd.Stderr = output

		err := proxyCmd.Start()
		Expect(err).ToNot(HaveOccurred())

		Eventually(output).Should(gbytes.Say("Starting Fab3 on port 5000"))
	})

})
