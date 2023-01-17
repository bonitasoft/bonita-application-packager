# Bonita Application Bundle packager tool

Bonita Application Bundle packager tool is a tool provided by Bonitasoft to allow
Bonita users and customers to build the self-contained Bonita Applications.


## Pre-requisites

* Have Java JDK 11 installed and configured (available in default PATH)
* Have an internet connection available


## What the packager tool does

* Downloads the Bonita Tomcat bundle from the internet
* Extracts the bundle
* Adds your Bonita Application contained in `my-application/` folder
* Re-packages the Bonita Tomcat bundle with your Bonita Application inside (under a new name)


## Using the packager tool

* [Clone](https://docs.github.com/en/repositories/creating-and-managing-repositories/cloning-a-repository) this GitHub repository
* Put you Bonita artifacts in `my-application/` folder
* run `./mvnw package` of `mvwn.bat package`


## Specify a precise Bonita version

Pass Maven extra parameter `bonita.branding.version`.

Eg. :
```shell
./mvnw package -Dbonita.branding.version=2022.2-u0
```


## Licensing

This repository is provided under [Gnu General Public License v2.0](LICENSE)