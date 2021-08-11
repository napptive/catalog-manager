/**
 * Copyright 2021 Napptive
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package metadata

import (
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	"syreclabs.com/go/faker"
)

func RunTests(provider MetadataProvider) {

	ginkgo.Context("Adding application metadata", func() {
		ginkgo.It("Should be able to add an application metadata", func() {
			app := utils.CreateTestApplicationInfo()

			returned, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())

		})

		ginkgo.It("Should be able to add an application metadata twice (update)", func() {
			app := utils.CreateTestApplicationInfo()

			_, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())

			_, err = provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())
		})

	})

	ginkgo.Context("Getting application metadata", func() {

		ginkgo.It("Should be able to get an application metadata", func() {
			app := utils.CreateTestApplicationInfo()

			returned, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())

			retrieved, err := provider.Get(&entities.ApplicationID{
				Namespace:       returned.Namespace,
				ApplicationName: returned.ApplicationName,
				Tag:             returned.Tag,
			})
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
			gomega.Expect(*retrieved).Should(gomega.Equal(*app))
		})

		ginkgo.It("Should not be able to get an application metadata if it does not exist", func() {
			_, err := provider.Get(&entities.ApplicationID{
				Namespace:       "repoTest",
				ApplicationName: "applName",
				Tag:             "",
			})
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})

	})

	ginkgo.Context("Removing application metadata", func() {
		ginkgo.It("Should be able to delete an application metadata", func() {

			for i := 0; i < 10; i++ {
				log.Debug().Int("i", i).Msg("...")
				app := utils.CreateTestApplicationInfo()

				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())

				err = provider.Remove(&entities.ApplicationID{
					Namespace:       returned.Namespace,
					ApplicationName: returned.ApplicationName,
					Tag:             returned.Tag,
				})
				gomega.Expect(err).Should(gomega.Succeed())
			}
		})
		ginkgo.It("Should not be able to delete an application metadata if it does not exist", func() {
			app := utils.CreateTestApplicationInfo()

			err := provider.Remove(&entities.ApplicationID{
				Namespace:       app.Namespace,
				ApplicationName: app.ApplicationName,
				Tag:             app.Tag,
			})
			gomega.Expect(err).ShouldNot(gomega.Succeed())

		})
	})

	ginkgo.Context("Getting an application metadata", func() {
		ginkgo.It("should be able to get an application metadata", func() {
			app := utils.CreateTestApplicationInfo()

			returned, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())

			retrieved, err := provider.Get(&entities.ApplicationID{
				Namespace:       app.Namespace,
				ApplicationName: app.ApplicationName,
				Tag:             app.Tag,
			})
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
			gomega.Expect(returned.CatalogID).Should(gomega.Equal(retrieved.CatalogID))
		})
		ginkgo.It("should not be able to return a non existing application metadata", func() {
			app := utils.CreateTestApplicationInfo()

			_, err := provider.Get(&entities.ApplicationID{
				Namespace:       app.Namespace,
				ApplicationName: app.ApplicationName,
				Tag:             app.Tag,
			})
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})

	})

	ginkgo.Context("Listing applications", func() {
		ginkgo.It("Should be able to list applications", func() {
			app := utils.CreateTestApplicationInfo()

			for i := 0; i < 5; i++ {
				app.Tag = faker.App().Version()
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}

			listRetrieved, err := provider.List("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(listRetrieved).ShouldNot(gomega.BeEmpty())
			gomega.Expect(len(listRetrieved)).Should(gomega.Equal(5))

		})
		ginkgo.It("Should be able to list an empty list of applications", func() {
			listRetrieved, err := provider.List("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(listRetrieved).Should(gomega.BeEmpty())
		})
		ginkgo.It("Should be able to list applications in a namespace", func() {
			numApps := 10
			targetNamespace := "target"
			for i := 0; i < numApps; i++ {
				app := utils.CreateTestApplicationInfo()
				if i%2 == 0 {
					app.Namespace = targetNamespace
				}
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}
			listRetrieved, err := provider.List(targetNamespace)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(len(listRetrieved)).Should(gomega.Equal(numApps / 2))
			for _, retrievedApp := range listRetrieved {
				gomega.Expect(retrievedApp.Namespace).Should(gomega.Equal(targetNamespace))
			}
		})
	})

	ginkgo.Context("Listing application summary", func() {
		ginkgo.It("Should be able to list applications", func() {
			namespace := "Namespace"
			appName := "App"

			for i := 0; i < 15; i++ {
				app := utils.CreateTestApplicationInfo()
				if i%3 == 0 {
					app.Namespace = namespace
				}
				if i%2 == 0 {
					app.ApplicationName = appName
				}
				app.Tag = faker.App().Version()
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}

			listRetrieved, err := provider.ListSummary("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(listRetrieved).ShouldNot(gomega.BeEmpty())

		})
		ginkgo.It("Should be able to list applications without logo in metadata", func() {

			var app *entities.ApplicationInfo
			for i := 0; i < 15; i++ {
				if i%3 == 0 {
					app = utils.CreateTestApplicationInfo()
				} else {
					app = utils.CreateTestApplicationInfoWithoutLogo()
				}
				app.Tag = faker.App().Version()
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}

			listRetrieved, err := provider.ListSummary("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(listRetrieved).ShouldNot(gomega.BeEmpty())

		})
		ginkgo.It("Should be able to list an empty list of applications", func() {
			listRetrieved, err := provider.ListSummary("")
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(listRetrieved).Should(gomega.BeEmpty())
		})
	})

	ginkgo.Context("Getting summary", func() {
		ginkgo.It("should be able to get summary", func() {
			numApp := 5
			for i := 0; i < numApp; i++ {
				app := utils.CreateTestApplicationInfo()
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}
			// Fill cache
			provider.(*ElasticProvider).FillCache()

			summary, err := provider.GetSummary()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(summary).ShouldNot(gomega.BeNil())
			gomega.Expect(summary.NumNamespaces).Should(gomega.Equal(numApp))
			gomega.Expect(summary.NumApplications).Should(gomega.Equal(numApp))
			gomega.Expect(summary.NumTags).Should(gomega.Equal(numApp))

		})
		ginkgo.It("should be able to get summary", func() {
			numApp := 5
			namespace1 := "namespace1"
			namespace2 := "namespace2"
			for i := 1; i <= numApp; i++ {
				app := utils.CreateTestApplicationInfo()
				if i%2 == 0 {
					app.Namespace = namespace1
				} else {
					app.Namespace = namespace2
				}
				returned, err := provider.Add(app)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())
			}
			// Fill cache
			provider.(*ElasticProvider).FillCache()

			summary, err := provider.GetSummary()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(summary).ShouldNot(gomega.BeNil())
			gomega.Expect(summary.NumNamespaces).Should(gomega.Equal(2))
			gomega.Expect(summary.NumApplications).Should(gomega.Equal(numApp))
			gomega.Expect(summary.NumTags).Should(gomega.Equal(numApp))

		})
	})

	ginkgo.Context("Checking if an application exists", func() {
		ginkgo.It("Should be able to check if an application exists", func() {
			app := utils.CreateTestApplicationInfo()

			returned, err := provider.Add(app)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(returned.CatalogID).ShouldNot(gomega.BeEmpty())

			exists, err := provider.Exists(&entities.ApplicationID{
				Namespace:       app.Namespace,
				ApplicationName: app.ApplicationName,
				Tag:             app.Tag,
			})
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(exists).Should(gomega.BeTrue())

		})
		ginkgo.It("Should be able to check if when an application does not exist", func() {
			app := utils.CreateTestApplicationInfo()

			exists, err := provider.Exists(&entities.ApplicationID{
				Namespace:       app.Namespace,
				ApplicationName: app.ApplicationName,
				Tag:             app.Tag,
			})
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(exists).ShouldNot(gomega.BeTrue())

		})
	})

}
