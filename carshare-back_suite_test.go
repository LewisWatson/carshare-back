package main_test

import (
	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var api *api2go.API

func TestCarShareBack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Car Share API Suite")
}
