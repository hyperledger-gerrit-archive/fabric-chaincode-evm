package addressgenerator_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAddressgenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Addressgenerator Suite")
}
