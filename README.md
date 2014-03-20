PAT (Performance Acceptance Tests)
==================================
The goal of this project is to create a super-simple load generation/performance testing framework for quickly and easily running load against Cloud Foundry. The tool has both a command line UI, for running quickly during a build and a web UI for tracking longer-running tests.


Running PAT as a Cloud Foundry App
==================================
Ensure your Cloud Foundry version is current and running

1) Clone the project

        git clone https://github.com/cloudfoundry-community/pat

2) Push the project to Cloud Foundry with our 'go' buildpack that adds gocart support.

        cf push -b https://github.com/jberkhahn/cloudfoundry-buildpack-go pat

3) Open the browser and go to the provided URL



Setting up PAT to run locally
==================================
To setup this project, a number of requires need to be met for GO.

1) Ensure that GO1.2 (64bit version) has been installed on the system.

2) Setup the GOPATH

        export GOPATH=~/go (or any other workspace repository for all your go code)

        export PATH=$GOPATH/bin:$PATH

3) Install [gocart] (https://github.com/vito/gocart)

        go get github.com/vito/gocart

4) Clone the project to the correct location

        mkdir -p $GOPATH/src/github.com/cloudfoundry-community

        cd $GOPATH/src/github.com/cloudfoundry-community

        git clone https://github.com/cloudfoundry-community/pat

        cd pat

        gocart

5) Install [gcf] (https://github.com/cloudfoundry/cli)

Develop
===================================
To develop for this project, you will first need to go through the "Setting up PAT" section. This project will
be maintained through the use of standard pull requests. When issuing a pull request, make sure to include sufficient
testing through the ginkgo package (see below) to go along with any code changes. The tests should document 
functionality and provide an example of how to use it.  

1) Go through the "Setting up PAT to run locally" section

2) Install [ginkgo] (http://onsi.github.io/ginkgo/#getting_ginkgo)

        go install github.com/onsi/ginkgo/ginkgo

3) Write and test your code following the ginkgo standards

4) Run all tests within the repository

        ginkgo -r

Run
==================================
If you wish to run PAT as a Cloud Foundry app, please refer to the section in the beginning of this guide.

There are three ways to run PAT locally. For all three ways, you must first:

1) Go through the "Setting up PAT to run locally" section

2) Make sure that you have targeted a Cloud Foundry environment using the gcf tool (# gcf login)

### Option 1. Run the source code directly

1) Change into the top level of this project

        cd $GOPATH/src/github.com/cloudfoundry-community/pat

2) Execute the command line

        go run main.go -workload=gcf:push

### Option 2. Compile and run an executable

1) Change into the top level of this project

        cd $GOPATH/src/github.com/cloudfoundry-community/pat
        go install

2) Run the PAT executable from the command line

        pat -workload=gcf:push

### Option 3. Run PAT with a web user interface

1) Change into the top level of this project

        cd $GOPATH/src/github.com/cloudfoundry-community/pat

2) Run PAT selecting the HTTP server option

        go run main.go -server

3) Open a browser and go to <http://localhost:8080/ui>


### Example command-line usage (using option 2 to illustrate):

    pat -h   # will output all of the command line options if installed the recommended way

    pat -concurrency=5 -iterations=5  # This will start 5 concurrent threads all pushing 1 application

    pat -concurrency=5 -iterations=1  # This will only spawn a single concurrent thread instead of the 5 you requested because you are only pushing a single application

    pat -silent  # If you don't want all the fancy output to be shown (results can be found in a CSV)

    pat -workload=gcf:push,gcf:push,..  # Select the workload operations you want to run (See "Workload options" below)

    pat -workload=dummy  # Run the tool with a dummy operation (not against a CF environment)

    pat -config=config/template.yml  # Include a configuration template specifying any number of command line arguments. (See "Using a Configuration file" section below).

    pat -rest:target http://api.xyz.abc.net \
        -rest:username=testuser1@xyz.com \
        -rest:password=PASSWORD \
        -rest:space=xyz_space  \
        -workload=rest:target,rest:login,rest:push,rest:push \
        -concurrency=5 -iterations=20 -interval=10 # Use the REST API to make operation requests instead of gcf 

### Workload options
The `workload` option specified a comma-separated list of workloads to be used in the test.
The following options are available:

- `rest:target` - sets the CF target. Mandatory to include before any other rest operations are listed.
- `rest:login` - performs a login to the REST api. This option requires `rest:target` to be included in the list of workloads.
- `rest:push` - pushes a simple Ruby application using the REST api. This option requires both `rest:target` and `rest:login` to be included in the list of workloads.
- `gcf:push` - pushes a simple Ruby application using the CF command-line
- `dummy` - an empty workload that can be used when a CF environment is not available.
- `dummyWithErrors` - an empty workload that generates errors. This can be used when a CF environment is not available.


Using a Configuration file
=====================================
PAT offers the ability to configure your command line arguments using a configuration file. There is an example in the root of the project
directory called config-template.yml. To use your own custom yaml configuration file, provide the full path to the 
configuration file. Any setting specified as a command line argument overrides the equivalent setting contained in the config file.

Example:
  
      pat -config=config-template.yml -iterations=2 # set iterations to 2 overriding whatever the config file says


Known Limitations / TODOs etc.
=====================================
 - Numerous :)
 - Unlikely to support Windows/Internet Explorer (certainly hasn't been tested on them)
 - Current feature set is a first-pass to get to something useful, contributions very welcome
 - Lots of stuff kept in memory and not flushed out
 - Creates lots of apps, does not delete them. We normally make sure we're targetted at a 'pats' space and just cf delete-space the space after to get rid of everything.
 - Only supports basic operations so far (push via gcf, target + login + push via rest api)
 - GCF workloads assume single already-logged-in-and-targetted user


