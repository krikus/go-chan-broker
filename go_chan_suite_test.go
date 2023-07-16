package gochanbroker_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGoChan(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GoChan Suite")
}
