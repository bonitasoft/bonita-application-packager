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

* Clone this GitHub repository
* Put you Bonita artifacts in `my-application/` folder
* run `mvnw package` of `mvwn.bat package`
