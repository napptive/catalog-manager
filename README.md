# catalog-manager
Napptive catalog-manager

Catalog manager is the component responsible for catalog management

## Layout structure

The layout structure is based on the default golang-template layout.

https://github.com/golang-standards/project-layout

## Usage

A make file is provided with the following targets:

* clean: Remove build files
* test: Run the available tests
* build: Build the files for your local environment
* build-darwin: Build the files for MacOS
* build-linux: Build the files for Linux
* k8s: Generate the Kubernetes deployment files
* docker-prep: Prepare the Dockerfile folder with all the extra files
* docker-build: Build the Dockerfile locally
* docker-push: Push the image to the selected repository. You must make login before to push the docker image.

---
**Important**

If you are developing with MacOS/Darwin, you must install gnu-sed.

```
brew install gnu-sed
```
---

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

![Check changes in the Main branch](https://github.com/napptive/catalog-manager/workflows/Check%20changes%20in%20the%20Main%20branch/badge.svg)

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
