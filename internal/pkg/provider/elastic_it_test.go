package provider

import (
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"strings"
	"syreclabs.com/go/faker"
	"time"
)

var mapping = `{
    "mappings": {
        "properties": {
          "id":         		{ "type": "keyword" },
          "Url":  				{ "type": "keyword" },
          "Repository":  		{ "type": "keyword" },
          "ApplicationName":	{ "type": "keyword" },
          "Tag":         		{ "type": "keyword" },
          "Readme": 			{ "type": "text" },
          "Metadata":  			{ "type": "text" }
      }
    }
}`

var _ = ginkgo.Describe("Elastic Provider test", func() {

	if !utils.RunIntegrationTests("provider") {
		log.Warn().Msg("elastic provider tests are skipped")
		return
	}

	index := strings.ToLower(faker.App().Name())
	log.Debug().Str("index", index).Msg("Elastic index")
	provider, err := NewElasticProvider(index, "http://localhost:9200")
	gomega.Expect(err).Should(gomega.Succeed())

	ginkgo.BeforeSuite(func() {
		err := provider.CreateIndex(mapping)
		gomega.Expect(err).Should(gomega.Succeed())
	})
	ginkgo.AfterSuite(func() {
		err := provider.DeleteIndex()
		gomega.Expect(err).Should(gomega.Succeed())
	})

	RunTests(provider)

	ginkgo.FIt("Getting an application metadata by ID", func() {
		app := utils.CreateApplicationMetadata()

		err := provider.Add(*app)
		gomega.Expect(err).Should(gomega.Succeed())

		// wait to be stored
		time.Sleep(time.Second * 2)

		retrieved, err := provider.GetByID(entities.ApplicationID{
			Url:             app.Url,
			Repository:      app.Repository,
			ApplicationName: app.ApplicationName,
			Tag:             app.Tag,
		})
		gomega.Expect(err).Should(gomega.Succeed())
		gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
		gomega.Expect(*retrieved).Should(gomega.Equal(*app))
	})

})
