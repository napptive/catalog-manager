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

/*
1.- environment variable:
RUN_INTEGRATION_TEST=all
RUN_INTEGRATION_TEST=userprovider

2.- run a postgres container
docker run -d --name local-postgres -e POSTGRES_PASSWORD=Pass2020 -p 5432:5432 postgres:13-alpine
docker exec -it local-postgres psql -h localhost -U postgres -d postgres -p 5432

3.- create the schema and table
CREATE SCHEMA IF NOT EXISTS catalog;
CREATE TABLE IF NOT EXISTS catalog.users (
  username VARCHAR(50) PRIMARY KEY NOT NULL,
  salt VARCHAR(16) NOT NULL,
  salted_password VARCHAR(256)
);

*/
package user_provider

import (
	"time"

	"github.com/napptive/catalog-manager/internal/pkg/utils"
	"github.com/napptive/rdbms/pkg/rdbms"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"

	"context"
)

var _ = ginkgo.Describe("Provider test", func() {

	if !utils.RunIntegrationTests("userprovider") {
		log.Warn().Msg("user provider tests are skipped")
		return
	}

	var connString = "host=localhost user=postgres password=Pass2020 port=5432"

	var provider UserProvider
	conn, err := rdbms.NewRDBMS().PoolConnect(context.Background(), connString)
	gomega.Expect(err).Should(gomega.Succeed())

	provider = NewUserProvider(conn, time.Second*25)
	gomega.Expect(provider).ShouldNot(gomega.BeNil())

	// empty the table
	ginkgo.AfterEach(func() {
		list, err := provider.List()
		gomega.Expect(err).Should(gomega.Succeed())

		for _, user := range list {
			err = provider.Remove(user.Username)
			gomega.Expect(err).Should(gomega.Succeed())
		}
	})

	RunTests(provider)

})
