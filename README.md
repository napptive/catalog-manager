# catalog-manager

The NAPPTIVE Catalog manager is the component responsible for providing a centralized repository to store application specifications aiming to facilitate the distribution and reusability of Cloud Native applications. Applications are expected to contain [OAM](https://oam.dev) entites with the addition of a new metadata entity to improve the user experience when navigating through the catalog.

## Getting started

* [Using the catalog as part of the NAPPTIVE Playground](https://docs.napptive.com/Catalog.html) (SaaS)
* [Hosting your own private catalog](docs/guides/PrivateCatalog.md)
  * [Kind](docs/guides/PrivateCatalogOnKind.md)
* [First steps](docs/guides/FirstSteps.md)

## Development

To run unit tests use:

```
make test
```

To run integration tests, first launch a docker container with elasticsearch and postgresql.

```
docker run -d --name elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" elasticsearch:7.11.2
docker run -d --name local-postgres -e POSTGRES_PASSWORD=Pass2020 -p 5432:5432 postgres:13-alpine
```

Next, you need to preload the database table definitions:

```
docker exec -it local-postgres psql -h localhost -U postgres -d postgres -p 5432

CREATE SCHEMA IF NOT EXISTS catalog;
CREATE TABLE IF NOT EXISTS catalog.users (
  username VARCHAR(50) PRIMARY KEY NOT NULL,
  salt VARCHAR(16) NOT NULL,
  salted_password VARCHAR(256)
);
```

then execute:

```
RUN_INTEGRATION_TEST=all make test
```

## Integration with Github Actions

This repository is integrated with GitHub Actions.

<a href="https://github.com/napptive/catalog-manager">![Check changes in the Main branch](https://github.com/napptive/catalog-manager/workflows/Check%20changes%20in%20the%20Main%20branch/badge.svg)
</a>

## Contributing

Let's make the catalog better, check our [contribution](./CONTRIBUTING.md) guidelines to discover how you can contribute to the project.

## License

 Copyright 2020 Napptive

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
