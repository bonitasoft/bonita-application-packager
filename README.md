# Bonita Application Bundle packager tool

Bonita Application Bundle packager tool is provided by Bonitasoft to allow
Bonita users and customers to build their self-contained Bonita Applications.


## Pre-requisites

* Have Java JDK 11 installed and configured (available in default PATH)
* Have an internet connection available
* [SUBSCRIPTION only] Having Maven configured to access Bonita Artifact Repository


## What the packager tool does

* Downloads the Bonita Tomcat bundle from the internet
* Extracts the bundle
* Adds your Bonita Application contained in `my-application/` folder
* Re-packages the Bonita Tomcat bundle with your Bonita Application inside (under a new name)


## Using the packager tool

### Community edition

* [Clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) this GitHub repository
* Put your Bonita artifacts in `community/my-application/` folder
* Run `./mvnw package -f community` (Unix / MacOS) or `mvwn.cmd package -f community` (Windows)

### Subscription edition

* [Clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) this GitHub repository
* Put your Bonita artifacts in `subscription/my-application/` folder
* Run `./mvnw package -f subscription` (Unix / MacOS) or `mvwn.cmd package -f subscription` (Windows)


## Specify a precise Bonita version to use

### Community edition

Pass Maven extra parameter `bonita.branding.version`.

Eg. :
```shell
./mvnw package -f community -Dbonita.branding.version=2023.1-u0
```

### Subscription edition

Pass Maven extra parameters `bonita.branding.version` and `bonita.tech.version`.

> ___Note___: parameters `bonita.branding.version` and `bonita.tech.version` must match precisely.
> 
> See [technical Id / platform version mapping list](https://documentation.bonitasoft.com/bonita/latest/version-update/product-versioning#_technical_id) for correspondence.

Eg. :
```shell
./mvnw package -f subscription -Dbonita.branding.version=2023.1-u1 -Dbonita.tech.version=8.0.1
```


## Licensing

This repository is provided under [Gnu General Public License v2.0](LICENSE)