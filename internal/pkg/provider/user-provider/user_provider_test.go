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
package user_provider

import (
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"syreclabs.com/go/faker"
)

func RunTests(provider UserProvider) {

	ginkgo.Context("Adding users", func() {
		ginkgo.It("Should be able to add a user", func() {

			user, err := provider.Add(&entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			})
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(user).ShouldNot(gomega.BeNil())
		})
		ginkgo.It("should not be able to add a user twice", func() {
			user := &entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			}
			retrieved, err := provider.Add(user)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(retrieved).ShouldNot(gomega.BeNil())

			_, err = provider.Add(user)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})

	ginkgo.Context("Removing users", func() {
		ginkgo.It("Should be able to delete a user", func() {
			user := &entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			}

			retrieved, err := provider.Add(user)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(retrieved).ShouldNot(gomega.BeNil())

			err = provider.Remove(user.Username)
			gomega.Expect(err).Should(gomega.Succeed())

		})
		ginkgo.It("Should not be able to delete a user if it does not exit", func() {
			user := &entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			}

			err := provider.Remove(user.Username)
			gomega.Expect(err).ShouldNot(gomega.Succeed())
		})
	})

	ginkgo.Context("Getting users", func() {
		ginkgo.It("Should be able to get a user", func() {
			user := &entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			}
			_, err := provider.Add(user)
			gomega.Expect(err).Should(gomega.Succeed())

			retrieved, err := provider.Get(user.Username)
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
			gomega.Expect(retrieved.Username).Should(gomega.Equal(user.Username))
			gomega.Expect(retrieved.Salt).Should(gomega.Equal(user.Salt))
			gomega.Expect(retrieved.SaltedPassword).Should(gomega.Equal(user.SaltedPassword))

		})
		ginkgo.It("Should not be able to get a non existing user", func() {
			user := &entities.User{
				Username:       faker.Internet().UserName(),
				Salt:           faker.Lorem().Characters(10),
				SaltedPassword: []byte(faker.Lorem().Characters(24)),
			}

			_, err := provider.Get(user.Username)
			gomega.Expect(err).ShouldNot(gomega.Succeed())

		})
	})

	ginkgo.Context("Listing users", func() {
		ginkgo.It("Should be able to return a list of users", func() {
			numUsers := 5
			for i := 0; i < numUsers; i++ {
				user := &entities.User{
					Username:       faker.Internet().UserName(),
					Salt:           faker.Lorem().Characters(10),
					SaltedPassword: []byte((faker.Lorem().Characters(24))),
				}
				retrieved, err := provider.Add(user)
				gomega.Expect(err).Should(gomega.Succeed())
				gomega.Expect(retrieved).ShouldNot(gomega.BeNil())
			}
			users, err := provider.List()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(users).ShouldNot(gomega.BeEmpty())

		})

		ginkgo.It("Should be able to return an empty list of users", func() {
			users, err := provider.List()
			gomega.Expect(err).Should(gomega.Succeed())
			gomega.Expect(users).Should(gomega.BeEmpty())
		})
	})


}
