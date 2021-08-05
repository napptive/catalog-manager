package metadata

import (
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"strings"
	"syreclabs.com/go/faker"
)

var _ = ginkgo.Describe("Elastic Provider test", func() {

	if !utils.RunIntegrationTests("provider") {
		log.Warn().Msg("elastic provider tests are skipped")
		return
	}

	index := strings.ToLower(faker.App().Name())
	index = strings.Replace(index, " ", "", -1)
	log.Debug().Str("index", index).Msg("Elastic index")
	provider, err := NewElasticProvider(index, "http://localhost:9200")
	gomega.Expect(err).Should(gomega.Succeed())

	ginkgo.BeforeEach(func() {
		err := provider.Init()
		gomega.Expect(err).Should(gomega.Succeed())
	})
	ginkgo.AfterEach(func() {
		err := provider.DeleteIndex()
		gomega.Expect(err).Should(gomega.Succeed())
	})

	RunTests(provider)

})
