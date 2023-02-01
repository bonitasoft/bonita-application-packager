# Bonita Application Bundle packager tool

Bonita Application Bundle packager tool is provided by Bonitasoft to allow
Bonita users and customers to build their self-contained Bonita Applications.


## Pre-requisites

* Have Java JDK 11 installed and configured (available in default PATH)
* Have an internet connection available
* [SUBSCRIPTION only] Having Maven configured to access [Bonita Artifact Repository](https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access)


## What the packager tool does

* Downloads the Bonita Tomcat bundle from the internet
* Extracts the bundle
* Adds your Bonita Application contained in `my-application/` folder
* Re-packages the Bonita Tomcat bundle with your Bonita Application inside (under a new name)


## Using the packager tool

### Community edition

* [Clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) this GitHub repository
* Put your Bonita artifacts in `community/my-application/` folder
* Run the command bellow with the Bonita version to use. The version must follow the [branding format](https://documentation.bonitasoft.com/bonita/latest/version-update/product-versioning#_technical_id) (e.g. `2023.1-u0`).
  * Unix / MacOS: `./mvnw package -f community -Dbonita.branding.version=<version>`
  * Windows: `mvwn.cmd package -f community -Dbonita.branding.version=<version>`
* If you are behind a proxy, pass the following parameters: `-Dhttp.proxyHost=proxy -Dhttp.proxyPort=8080`

### Subscription edition

* [Clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) this GitHub repository
* Put your Bonita artifacts in `subscription/my-application/` folder
* Run `./mvnw package -f subscription` (Unix / MacOS) or `mvwn.cmd package -f subscription` (Windows)
* Run the command bellow with the Bonita version to use. The version must follow the [technical format](https://documentation.bonitasoft.com/bonita/latest/version-update/product-versioning#_technical_id) (e.g. `8.0.0`).
    * Unix / MacOS: `./mvnw package -f subscription -Dbonita.tech.version=<version>`
    * Windows: `mvwn.cmd package -f subscription -Dbonita.tech.version=<version>`
* If you are behind a proxy, no need to pass extra parameters. Your Maven local settings will be used instead.


## Licensing

This repository is provided under [GNU General Public License v2.0](LICENSE).
