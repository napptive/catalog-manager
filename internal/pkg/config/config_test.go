package config

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Config tests", func() {

	cfg := Config{}

	ginkgo.It("Should be valid", func() {
		gomega.Expect(cfg.IsValid()).To(gomega.Succeed())
	})

})