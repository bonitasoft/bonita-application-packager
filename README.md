# (DEPRECATED) Bonita Application Packager tool

**DEPRECATED: This experimental tool can only be used with Bonita version 2023.1. Starting from 2023.2, the packaging process is replaced by a [Maven plugin](https://github.com/bonitasoft/bonita-project-maven-plugin).**

The Bonita Application Packager tool is provided by Bonitasoft to allow you to build a Bonita Tomcat bundle or a Bonita Docker image containing your custom application.

The advantage of such a packaging is that Bonita automatically installs your entire application (pages, Rest API extensions, themes, BDM, ...) at startup, relieving you from having to install all artifacts by hand one by one.

The tool can be used by both Community and Subscription users.


## Pre-requisites

* Systems compatible to execute this tool are:
    * Linux x86-64
    * Microsoft Windows x86-64
    * Apple macOS x86-64
    * Apple macOS ARM64
* Minimum Bonita version is **2023.1**
* The custom application is a ZIP file containing all Bonita deployable artifacts, such as:
    * BDM (Business Data Model)
    * BDM Access Control **(Subscription only)**
    * Layouts
    * Living Applications
    * Organizations
    * Pages
    * Processes
    * Profiles **(Subscription only)**
    * REST API Extensions
    * Themes
* **(Subscription only)** The *Application Archive* generated with [Bonita Continuous Delivery](https://documentation.bonitasoft.com/bcd/latest/livingapp_build) is a proper candidate for a custom application.
* **(Subscription only)** The *Application Configuration Archive* (*.bconf* file) generated with [Bonita Continuous Delivery](https://documentation.bonitasoft.com/bcd/latest/livingapp_build) is an extra deployable artifact compatible with this tool.
* The custom application is located and accessible on your filesystem.
* The custom application has been built for Bonita 2023.1+ (Bonita runtime version 8.0+).

### Bonita Tomcat bundle usage

* Have downloaded on your filesystem a Bonita Tomcat bundle compatible with your custom application.

### Bonita Docker image usage

* Have Docker client installed and configured.
* Have an internet connection available.
* **(Subscription only)** Have credentials to access [Bonita Artifact Repository](https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access)


## How to use the packager tool

1. Download the executable binary from the [GitHub Releases page](https://github.com/bonitasoft/bonita-application-packager/releases). Choose a binary compatible with your system (see above pre-requisites).
1. Execute the binary in a terminal.
1. (Optional) To make the binary executable anywhere, copy the binary in a folder already defined in your `PATH` environment variable or update the `PATH` environment variable with the folder containing the binary.


**Usage:**

```shell
bonita package [tomcat|docker] [OPTIONS] PATH_TO_APPLICATION_ZIP_FILE
```

**Options are:**

**For Tomcat**

```
  -b, --bonita-tomcat-bundle string   (Optional) Specify path to the Bonita tomcat bundle file (Bonita*.zip) used to build.
                                      If not passed, looking for a Bonita tomcat bundle in current folder
  -h, --help                          help for tomcat

Global Flags:
  -c, --configuration-file string   (Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)
  -v, --verbose                     More verbose information output
```


**For Docker**

```
  -i, --bonita-base-image string   Specify Bonita base docker image (default "bonita:latest")
  -h, --help                       help for docker
  -p, --registry-password string   Specify password to authenticate against Bonita base docker image Registry
                                   If --registry-username is provided and not --registry-password, password will be prompted interactively and never issued to the console
  -u, --registry-username string   Specify username to authenticate against Bonita base docker image Registry
  -t, --tag string                 Docker image tag to use when building (default "my-bonita-application:latest")

Global Flags:
  -c, --configuration-file string   (Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)
  -v, --verbose                     More verbose information output
```


## After packaging usage

### Tomcat bundle

The Tomcat output of the packager tool is like a usual Bonita Tomcat bundle. To use it, unzip it and execute `start-bonita[.sh|.bat]` script.

Read [Installation of Bonita Tomcat bundle](https://documentation.bonitasoft.com/bonita/latest/runtime/tomcat-bundle) for more information.


### Docker image

The Docker output of the packager tool is like a usual Bonita Docker image. To use it, run appropriate command:

**Community release:**

```shell
docker run --name my-bonita-app -d -p 8080:8080 <my-docker-image-app>
```

**Subscription release:**

```shell
docker run --name my-bonita-app -h <hostname> -v <license-folder>:/opt/bonita_lic/ -d -p 8080:8080 <my-docker-image-app>
```

Read [Deploy Bonita Runtime with Docker](https://documentation.bonitasoft.com/bonita/latest/runtime/bonita-docker-installation) for complete options on how to run a Bonita-based Docker container.


### Installation of Bonita Admin & User application pages

[Bonita Admin Application](https://documentation.bonitasoft.com/bonita/latest/runtime/admin-application-overview) and [Bonita User Application](https://documentation.bonitasoft.com/bonita/latest/runtime/user-application-overview) are no longer installed after the packaging process. If your custom application is using pages from one of those applications, a property must be set in order to install those pages at Bonita Runtime startup:

* Within a Tomcat bundle: set the Bonita runtime property `bonita.runtime.custom-application.install-provided-pages=true` in bundle configuration (see [Bonita Runtime Configuration](https://documentation.bonitasoft.com/bonita/latest/runtime/runtime-configuration-index) for more information)

* Within a Docker container: set the [environment variable](https://documentation.bonitasoft.com/bonita/latest/runtime/bonita-docker-installation#_environment_variables) `INSTALL_PROVIDED_PAGES=true` when running container

The concerned pages are:

* [Admin Process list page](https://documentation.bonitasoft.com/bonita/latest/runtime/admin-application-process-list)
* [Admin Case list page](https://documentation.bonitasoft.com/bonita/latest/runtime/cases)
* [Admin Task list page](https://documentation.bonitasoft.com/bonita/latest/runtime/admin-application-task-list)
* [User Process list page](https://documentation.bonitasoft.com/bonita/latest/runtime/user-process-list)
* [User Case list page](https://documentation.bonitasoft.com/bonita/latest/runtime/user-application-case-list)
* [User Case details page](https://documentation.bonitasoft.com/bonita/latest/runtime/user-application-case-list#case-details)
* [User Task list page](https://documentation.bonitasoft.com/bonita/latest/runtime/user-task-list)


## Examples

### Bonita Tomcat bundle usage

* Basic usage:

```shell
bonita package tomcat /path/to/my-custom-application.zip
```

The result is a ZIP file located under `output/` folder of the current folder.

If you do not specify the path to a Bonita Tomcat bundle file, the tool takes the first `Bonita*.zip` file located in the **current folder**.


* Specify the path to the Bonita Tomcat bundle:

```shell
bonita package tomcat --bonita-tomcat-bundle /path/to/BonitaCommunity-2023.1-u0.zip /path/to/my-custom-application.zip
```
or simpler
```shell
bonita package tomcat -b /path/to/BonitaCommunity-2023.1-u0.zip /path/to/my-custom-application.zip
```

Output: `./output/BonitaCommunity-2023.1-u0-application.zip`


* **(Subscription only)** Specify the path to the *Application Configuration*:

```shell
bonita package tomcat --bonita-tomcat-bundle /path/to/BonitaSubscription-2023.1-u0.zip --configuration-file /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```
or simpler
```shell
bonita package tomcat -b /path/to/BonitaSubscription-2023.1-u0.zip -c /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```

Output: `./output/BonitaSubscription-2023.1-u0-application.zip`


### Bonita Docker image usage

* Basic usage:

```shell
bonita package docker /path/to/my-custom-application.zip
```

The result is a Docker image tagged as `my-bonita-application:latest`.

By default, the Bonita base image used is `bonita:latest`, located on [DockerHub](https://hub.docker.com/_/bonita).


* Specify the Bonita base image:

```shell
bonita package docker --bonita-base-image my-registry/bonita:2023.1-u0 /path/to/my-custom-application.zip 
```
or simpler
```shell
bonita package docker -i my-registry/bonita:2023.1-u0 /path/to/my-custom-application.zip 
```

If the `my-registry` Docker registry requires authentication, provide `--registry-username` and `--registry-password` options.


* Specify the Docker image tag:

```shell
bonita package docker --tag my-docker-application:1.0.0 /path/to/my-custom-application.zip
```
or simpler
```shell
bonita package docker -t my-docker-application:1.0.0 /path/to/my-custom-application.zip
```

Be careful to respect [Docker image tag naming constraints](https://docs.docker.com/engine/reference/commandline/tag/)


* **(Subscription only)** Specify the path to the *Application Configuration*:

```shell
bonita package docker --configuration-file /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```
or simpler
```shell
bonita package docker -c /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```

The result is a Docker image containing both your custom application and its configuration.


* **(Subscription only)** Usage with Bonita Artifact Repository registry:

```shell
bonita package docker --bonita-base-image bonitasoft.jfrog.io/docker/bonita-subscription:8.0.0 --registry-username <access-login> --registry-password <access-token> /path/to/my-custom-application.zip
```
or simpler
```shell
bonita package docker -i bonitasoft.jfrog.io/docker/bonita-subscription:8.0.0 -u <access-login> -p <access-token> /path/to/my-custom-application.zip
```

If password is not provided, you will be prompted to enter it on the command line (it will not be issued in the console).

See [Bonita Artifact Repository documentation](https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access#credentials) on how to get your credentials.

The tool accepts all available version formats for Bonita base image. For example: `8.0`, `8.0.0`, `2023.1`, `2023.1-u0`.


## Display the version of this tool

Simply call:

`bonita version`

`bonita --version`

or

`bonita -v`


## Building

Run provided script:

`./build-all-os.sh`


## Updating 3rd-party dependencies

`go get -u ./... && go mod tidy`


## Contributing

If you want to contribute, ask questions about the project, report bug, see the [contributing guide](https://github.com/bonitasoft/bonita-developer-resources/blob/master/CONTRIBUTING.MD).


## Licensing

This repository is provided under [GNU General Public License v2.0](LICENSE).
