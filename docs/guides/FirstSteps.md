# First steps

Once the catalog is installed, you are ready to upload applications!.

The NAPPTIVE Catalog is highly inspired by the way container registries work. The applications uploaded by a user will be available to all the users who have access to the catalog.

## Applications

Each user will have a repository whose name will match their username. It will be in this repository where you must upload your applications.

An application in the catalog is a collection of YAML entities that specify which elements are involved in its deployment. In order to fully identify the application, a metadata file **is required**. This enables a better discovery of what is available on the catalog. To upload an application, please include a **metadata.yaml** file with the other elements of the application, with the following content:

```yaml
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
# Name of the application, not necessarily a valid k8s name.
name: "My App Name"
version: 1.0
description: Short description for searchs. Long one plus how to goes into the README.md
# Keywords facilitate searches on the catalog
keywords:
  - "tag1"
  - "tag2"
  - "tag3" 
license: "Apache License Version 2.0"
url: "https://..."
doc: "https://..."
# Requires gives a list of entities that are needed to launch the application.
requires:
  traits:
    - my.custom.trait
  scopes:
    - my.custom.scope
  # K8s lists Kubernetes specific entities. This provides a separation between OAM entities in an orchestration-agnostic environment, and applications that specifically require Kubernetes.
  k8s:
    - apiVersion: my.custom.package
      kind: CustomEntityKind
      name: name
# The logo can be used as visual information when listing the catalog so the user recognizes more easily the application.
logo:
  - src: "https://my.domain/path/logo.png"
    type: "image/png"
    size: "120x120"
```

As for the layout of the application YAML we recommend:

```text
/<application>
  |- metadata.yaml
  |- README.md
  |- app_cfg.yaml
  |
  |- <opt_directory>
  |  |- OAM_entity.yaml
  |  \- K8s_entity.yaml
  |- OAM_entity.yaml
  \- K8s_entity.yaml
```

Adding a `README.md` file will also help consumers of the application understand how it is intended to be deployed, how parameters interact, etc.

## CLI

There is a client to manage catalog operations. Visit [github repo](https://github.com/napptive/catalog-cli).

```bash
$ catalog --help
The catalog command provides a set of methods to interact with the Napptive Catalog

Usage:
  catalog [flags]
  catalog [command]

Examples:
$ catalog

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  info        Get the principal information of an application.
  list        List the applications
  pull        Pull an application from catalog.
  push        Push an application in the catalog.
  remove      Remove an application from catalog.
  summary     Get te catalog summary.
```

## Upload an application

A user will upload in his repository as many applications as he wants. The application name must comply with the following rule:

```text
repository/applicationName:[version]
```

To upload an application, execute:

```bash
$ catalog push <repo>/<appName>:<version> <application_path> --catalogAddress localhost --catalogPort 37060  --useTLS=false

STATUS     INFO
SUCCESS    <repositiry>/<appName> added to catalog
```

at this moment, the application is ready for everyone that wants to use it.

## Download an application

A user can download any available application by executing:

```bash
$ catalog pull <repo>/<appName>:<version> <application_path> --catalogAddress localhost --catalogPort 37060  --useTLS=false

STATUS     INFO
SUCCESS    application saved on ./<appName>.tgz
```

all the application files will be downloaded in _appName.tgz_ file.

## Remove an application

To remove an existing application, execute:

```bash
$ catalog remove <repo>/<appName>:<version> <application_path> --catalogAddress localhost --catalogPort 37060  --useTLS=false

STATUS     INFO
SUCCESS    <repo>/<appName>:<version> removed from catalog
```

## List avaiblable applications

To receive all the applications available in the cataglog, execute:

```bash
$ catalog --catalogAddress localhost --catalogPort 37060  --useTLS=false list
APPLICATION                 NAME
repository/appName:version  Application name
```

## Get Summary

To receive a summary of the number of applications available in the catalog:

```bash
$ catalog summary --catalogAddress localhost --catalogPort 37060  --useTLS=false
NAMESPACES    APPLICATIONS    TAGS
<num>         <num>           <num>
```

## Get application info

To get the information of an application, execute:

```bash
$ catalog  --catalogAddress localhost --catalogPort 37060  --useTLS=false  info <repo>/<appName>:<version>

APP_ID                      NAME
<repo>/<appName>:<version>  Application Name

DESCRIPTION
<application description>

README
<readme file>
```
