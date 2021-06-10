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

	ginkgo.FIt("checking concurrency", func() {
		/*for i := 0; i < 15; i++ {
			app := utils.CreateTestApplicationInfo()
			app.Tag = faker.App().Version()
			returned, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
		}
		 */
		addApp(provider, 0)
		time.Sleep(time.Second * 2)

		addApp(provider, 0)
		time.Sleep(time.Second * 1)
		addApp(provider, 0)
		addApp(provider, 0)



		provider.Finish()

		for i:=1; i<=20; i++ {
			if i%5 == 0 {
				//go addApp(provider, i)
			}else{
				//time.Sleep(time.Millisecond * 500)
				//go checkList(provider, i)
			}
		}

	})


})

func addApp(provider *ElasticProvider, thread int) {
	app := utils.CreateTestApplicationInfo()
	app.Tag = faker.App().Version()
	returned, err := provider.Add(app)
	gomega.Expect(err).Should(gomega.Succeed())
	gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
}

func checkList(provider *ElasticProvider, thread int){
	_, err := provider.ListSummary("")
	gomega.Expect(err).Should(gomega.Succeed())
	// gomega.Expect(listRetrieved).ShouldNot(gomega.BeEmpty())
}
