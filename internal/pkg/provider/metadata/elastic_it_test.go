package metadata

import (
	"strings"
	"time"

	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
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

	ginkgo.It("Getting an application metadata by Search Method", func() {
		app := utils.CreateTestApplicationInfo()

		returned, err := provider.Add(app)
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(returned).ShouldNot(gomega.BeNil())

		// wait to be stored
		time.Sleep(time.Second)

		retrieved, err := provider.SearchByApplicationID(entities.ApplicationID{
			Namespace:       returned.Namespace,
			ApplicationName: returned.ApplicationName,
			Tag:             returned.Tag,
		})
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
		gomega.Expect(*retrieved).Should(gomega.Equal(*app))
	})

})
