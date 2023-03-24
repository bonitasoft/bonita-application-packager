# Bonita Application Packager tool

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
* **(Subscription only)** DEPRECATED: Have credentials to access *Quay.io* ([see information on how to access Quay.io](https://customer.bonitasoft.com/download/request))
* **(Subscription only)** Have credentials to access [Bonita Artifact Repository](https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access)


## How to use the packager tool

1. Download the executable binary from the [GitHub Releases page](https://github.com/bonitasoft/bonita-application-packager/releases). Choose a binary compatible with your system (see above pre-requisites).
1. Execute the binary in a terminal.
1. (Optional) To make the binary executable anywhere, copy the binary in a folder already defined in your `PATH` environment variable or update the `PATH` environment variable with the folder containing the binary.


**Usage:**

```
bonita-packager [-tomcat|-docker] [OPTIONS] PATH_TO_APPLICATION_ZIP_FILE
```

**Options are:**

```
-bonita-base-image string
    Specify Bonita base docker image (default "bonita:latest")
-bonita-tomcat-bundle string
    (Optional) Specify path to the Bonita tomcat bundle file (Bonita*.zip) used to build
-configuration-file string
    (Optional) Specify path to the Bonita configuration file (.bconf) associated to your custom application (Subscription only)
-docker
    Choose to build a docker image containing your application,
    use -tag to specify the name of your built image
    use -bonita-base-image to specify a Bonita docker base image different from the default, which is 'bonita:latest'
    use -registry-username and -registry-password if you need to authenticate against the docker image registry to pull Bonita docker base image
-help
    Print complete usage of this tool
-registry-password string
    Specify password to authenticate against Bonita base docker image Registry
    If -registry-username is provided and not -registry-password, password will be prompted interactively and never issued to the console
-registry-username string
    Specify username to authenticate against Bonita base docker image Registry
-tag string
    Docker image tag to use when building (default "my-bonita-application:latest")
-tomcat
    Choose to build a Bonita Tomcat bundle containing your application
    use -bonita-tomcat-bundle to specify the path to the Bonita tomcat bundle file (Bonita*.zip); otherwise the file is looked for in the current folder
-verbose
    Enable verbose (debug) mode
-version
    Print tool version and exit
```


## Examples

### Bonita Tomcat bundle usage

* Basic usage:

```
bonita-packager -tomcat /path/to/my-custom-application.zip
```

The result is a ZIP file located under `output/` folder of the current folder.

If you do not specify the path to a Bonita Tomcat bundle file, the tool takes the first `Bonita*.zip` file located in the current folder.


* Specify the path to the Bonita Tomcat bundle:

```
bonita-packager -tomcat -bonita-tomcat-bundle /path/to/BonitaCommunity-2023.1-u0.zip /path/to/my-custom-application.zip
```

Output: `./output/BonitaCommunity-2023.1-u0-application.zip`


* **(Subscription only)** Specify the path to the *Application Configuration*:

```
bonita-packager -tomcat -bonita-tomcat-bundle /path/to/BonitaSubscription-2023.1-u0.zip -configuration-file /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```

Output: `./output/BonitaSubscription-2023.1-u0-application.zip`


### Bonita Docker image usage

* Basic usage:

```
bonita-packager -docker /path/to/my-custom-application.zip
```

The result is a Docker image tagged as `my-bonita-application:latest`.

By default, the Bonita base image used is `bonita:latest`, located on [DockerHub](https://hub.docker.com/_/bonita).


* Specify the Bonita base image:

```
bonita-packager -docker -bonita-base-image my-registry/bonita:2023.1-u0 /path/to/my-custom-application.zip 
```

If the `my-registry` Docker registry requires authentication, provide `-registry-username` and `-registry-password` options.


* Specify the Docker image tag:

```
bonita-packager -docker -tag my-docker-application:1.0.0 /path/to/my-custom-application.zip
```


* **(Subscription only)** Specify the path to the *Application Configuration*:

```
bonita-packager -docker -configuration-file /path/to/my-custom-application.bconf /path/to/my-custom-application.zip
```

The result is a Docker image containing both your custom application and its configuration.


* **(Subscription only)** Usage with Bonita Artifact Repository registry:

```
bonita-packager -docker -bonita-base-image bonitasoft.jfrog.io/docker/bonita-subscription:8.0.0 -registry-username <access-login> -registry-password <access-token> /path/to/my-custom-application.zip
```

See [Bonita Artifact Repository documentation](https://documentation.bonitasoft.com/bonita/latest/software-extensibility/bonita-repository-access#credentials) on how to get your credentials.

The tool accepts all available version formats for the base image. For example: `8.0`, `8.0.0`, `2023.1`, `2023.1-u0`.


* **(Subscription only)** Usage with Quay.io registry (DEPRECATED):

Quay.io is deprecated and is replaced by Bonita Artifact Repository.

```
bonita-packager -docker -bonita-base-image quay.io/bonitasoft/bonita-subscription:8.0.0 -registry-username <access-login> -registry-password <access-token> /path/to/my-custom-application.zip
```

See [information on how to get your credentials to Quay.io](https://customer.bonitasoft.com/download/request).

The tool accepts all available version formats for the base image. For example: `8.0`, `8.0.0`, `2023.1`, `2023.1-u0`.


## Contributing

If you want to contribute, ask questions about the project, report bug, see the [contributing guide](https://github.com/bonitasoft/bonita-developer-resources/blob/master/CONTRIBUTING.MD).


## Licensing

This repository is provided under [GNU General Public License v2.0](LICENSE).
