[![Build Status](https://travis-ci.org/cloudfoundry-incubator/pat.svg?branch=master)](https://travis-ci.org/cloudfoundry-incubator/pat)

PAT (Performance Acceptance Tests)
==================================
The goal of this project is to create a super-simple load generation/performance testing framework for quickly
and easily running load against Cloud Foundry. The tool has both a command line UI, for running quickly during
a build and a web UI for tracking longer-running tests.

To run PATs, you could download the binary executable we provide, or you could clone and run the repository if you want the latest version of PATs

Download PATs Binary
==================================
If you just want to run PATs, you could download our PATs binary file.

Note: It is important that `cf` is accessable on your `$PATH` if you intend to use any of the 'cf:' workloads (see [CF Cli](http://github.com/cloudfoundry/cli) for instructions on installing the cloudfoundry cli).

Available Binary:
- Mac OSx 64bit
- Linux 64bit
- Windows 64bit

Goto https://github.com/cloudfoundry-incubator/pat/releases to download



Clone and Setting up to run locally
==================================
These steps are to setup this project and have it run locally on your system. This includes a number of
requirements for Go and the dependent libraries. If you wish to only run this project as a Cloud Foundry
application, see the instructions on "Running PAT as a Cloud Foundry App" below.

1) Ensure that [Go](http://golang.org/) version 1.2.x-64bit has been installed on the system

2) Setup the GOPATH

    export GOPATH=~/go
    export PATH=$GOPATH/bin:$PATH

3) Install [gocart] (https://github.com/vito/gocart)

    go get github.com/vito/gocart

4) Download PAT and install the necessary dependencies

    go get github.com/cloudfoundry-incubator/pat
      *(Ignore any warnings about "no buildable Go source files")
      *(Ignore errors in src/github.com/cloudfoundry-incubator/pat/workloads/cf.go")
    cd $GOPATH/src/github.com/cloudfoundry-incubator/pat
    gocart

5) See [CF CLI] (https://github.com/cloudfoundry/cli) for instructions on installing `cf`

Note: It is important that `cf` is accessable on your `$PATH` if you intend to use any of the 'cf:' workloads.

Running PAT
=================================

## Running Locally

If you wish to run PAT as a Cloud Foundry app(work in progress), please refer to the section at the bottom of this page.

There are three ways to run PAT locally. For all three ways, you must first:

1) Go through the "Setting up PAT to run locally" section

2) Make sure that you have targeted a Cloud Foundry environment using the cf tool

    cf login

### Option 1. Run the source code directly
1) Change into the top level of this project

    cd $GOPATH/src/github.com/cloudfoundry-incubator/pat

2) Execute the command line

    go run main.go -workload=cf:push

### Option 2. Run the source code through a web interface

1) Change into the top level of this project

    cd $GOPATH/src/github.com/cloudfoundry-incubator/pat

2) Run PAT selecting the HTTP server option

    go run main.go -server

3) Open a browser and go to <http://localhost:8080/ui>

### Option 3. Compile and run a PAT executable

1) Change into the top level of this project

    cd $GOPATH/src/github.com/cloudfoundry-incubator/pat
    go install

2a) Run the PAT executable in command line mode

    pat -workload=cf:push

2b) Run the PAT executable in web interface mode

    pat -server

### Example command-line usage (using option 3 to illustrate):

    pat -h   # will output all of the command line options if installed the recommended way

    pat -concurrency=5 -iterations=5  # This will start 5 concurrent threads all pushing 1 application

    pat -concurrency=5 -iterations=1  # This will only spawn a single concurrent thread instead of the 5 you requested because you are only pushing a single application

    pat -concurrency=1..5 -concurrency:timeBetweenSteps=10  -iterations=5 # This will ramp from 1 to 5 workers, adding a worker every 10 seconds.

    pat -silent  # If you don't want all the fancy output to be shown (results can be found in a CSV)

    pat -list-workloads  # Lists the available workloads

    pat -workload=cf:push,cf:push,..  # Select the workload operations you want to run (See "Workload options" below)

    pat -workload=dummy  # Run the tool with a dummy operation (not against a CF environment)

    pat -config=config/template.yml  # Include a configuration template specifying any number of command line arguments. (See "Using a Configuration file" section below).

    pat -rest:target=http://api.xyz.abc.net \
        -rest:username=testuser1@xyz.com \
        -rest:password=PASSWORD \
        -rest:space=xyz_space  \
        -workload=rest:target,rest:login,rest:push,rest:push \
        -concurrency=5 -iterations=20 -interval=10 # Use the REST API to make operation requests instead of cf

### Workload options
The `workload` option specified a comma-separated list of workloads to be used in the test.
The following options are available:

- `rest:target` - sets the CF target. Mandatory to include before any other rest operations are listed.
- `rest:login` - performs a login to the REST api. This option requires `rest:target` to be included in the list of workloads.
- `rest:push` - pushes a simple Ruby application using the REST api. This option requires both `rest:target` and `rest:login` to be included in the list of workloads.
- `cf:push` - pushes an application using the CF command-line, defaults to pushing [Dora]("https://github.com/cloudfoundry/cf-acceptance-tests/tree/master/assets/dora").
- `dummy` - an empty workload that can be used when a CF environment is not available.
- `dummyWithErrors` - an empty workload that generates errors. This can be used when a CF environment is not available.

### Required arguments
Certain `workload` options require one or more arguments to be defined
The following are a list of arguments

- `-rest:target` - The Cloud Foundry URL PAT should target to. Mandatory if workload option `rest:target` is used.
- `-rest:username` - Username for workload option `rest:login`. PAT supports multi credentials, for example, if you supply  `-rest:username=user1,user2,user3`, PAT will loop through the list and use a different credential at each iteration. This argument is mandatory for workload option `rest:login`.
- `-rest:password` - Similar to `-rest:username`, used to define the password for workload option `rest:login`.

Using Redis to create a cluster of PAT workers
=====================================

Pat supports shipping workload to multiple instances using redis. This simple example starts four pat instances on the local computer which all communicate to run a workload.

    cd $GOPATH/src/github.com/cloudfoundry-incubator/pat
    redis-server redis/redis.conf # start up with in-memory only db config, good for testing, replace with a real config and change ports for real use
    VCAP_APP_PORT=8080 go run main.go -use-redis-worker=true -server -redis-port=63798 -redis-host=127.0.0.1 -redis-password=p4ssw0rd -use-redis-store # instance 1
    VCAP_APP_PORT=8081 go run main.go -use-redis-worker=true -server -redis-port=63798 -redis-host=127.0.0.1 -redis-password=p4ssw0rd -use-redis-store # instance 2
    VCAP_APP_PORT=8082 go run main.go -use-redis-worker=true -server -redis-port=63798 -redis-host=127.0.0.1 -redis-password=p4ssw0rd -use-redis-store # instance 3
    VCAP_APP_PORT=8083 go run main.go -use-redis-worker=true -server -redis-port=63798 -redis-host=127.0.0.1 -redis-password=p4ssw0rd -use-redis-store # instance 4


Using a Configuration file
=====================================
PAT offers the ability to configure your command line arguments using a configuration file. There is an example in the root of the project
directory called config-template.yml. To use your own custom yaml configuration file, provide the full path to the 
configuration file. Any setting specified as a command line argument overrides the equivalent setting contained in the config file.

Example:

    pat -config=config-template.yml -iterations=2 # set iterations to 2 overriding whatever the config file says

Error Codes
=====================================
In the event of an error during execution, the text of the error along with an error code will be returned to the user. Codes are as follows:

    10: Error parsing input
    20: Error in executing the workload

<!---
Running PATs as a Cloud Foundry App (In the works, some features might not work)
===================================

Ensure your Cloud Foundry version is current and running

1) Clone the project if you have not followed the "Setting up PAT to run locally" section

    git clone https://github.com/cloudfoundry-incubator/pat

2) Change into the PAT directory

3) Push the project to Cloud Foundry with our 'go' buildpack that adds gocart support

    cf push -b https://github.com/jberkhahn/cloudfoundry-buildpack-go pat

4) Open the browser and go to the provided URL
--->

Contributing
===================================
To contribute to this project, you will first need to go through the "Setting up PAT to run locally" section. This
project will be maintained through the use of standard pull requests. When issuing a pull request, make sure to
include sufficient testing through the ginkgo package (see below) to go along with any code changes. The tests
should document functionality and provide an example of how to use it.

1) Go through the "Setting up PAT to run locally" section

2) Install [ginkgo] (http://onsi.github.io/ginkgo/#getting_ginkgo)

        go install github.com/onsi/ginkgo/ginkgo

3) Write and test your code following the ginkgo standards

4) Install Prerequisites:

 - *Redis*: e.g. `brew install redis` (using [HomeBrew](https://github.com/Homebrew/homebrew) on OSX)

5) Run all tests within the repository

        ginkgo -r

Known Limitations / TODOs etc.
=====================================
 - Numerous :)
 - Unlikely to support Windows/Internet Explorer (certainly hasn't been tested on them)
 - Current feature set is a first-pass to get to something useful, contributions very welcome
 - Lots of stuff kept in memory and not flushed out
 - Creates lots of apps, does not delete them. We normally make sure we're targetted at a 'pats' space and just cf delete-space the space after to get rid of everything.
 - Only supports basic operations so far (push via cf, target + login + push via rest api)
 - cf workloads assume single already-logged-in-and-targetted user


