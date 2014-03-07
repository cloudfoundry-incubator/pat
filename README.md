PATs (Performance Acceptance Tests)
==================================
The goal of this project is to create a super-simple load generation testing framework for quickly and easily running load against Cloud Foundry.


Running PATs as a Cloud Foundry App
==================================
Ensure your Cloud Foundry version is current and running

1) Clone the project

        git clone https://github.com/julz/pat

2) Push the project to Cloud Foundry with the 'go' buildpack

        cf push -b https://github.com/jberkhahn/cloudfoundry-buildpack-go pat

3) Open the browser and go to the provided URL



Setting up PATs to run locally
==================================
To setup this project, a number of requires need to be met for GO.

1) Ensure that GO1.2 (64bit version) has been installed on the system.

2) Setup the GOPATH

        export GOPATH=~/go (or any other workspace repository for all your go code)

        export PATH=$GOPATH/bin:$PATH

3) Install [gocart] (https://github.com/vito/gocart)

        go get github.com/vito/gocart

4) Clone the project to the correct location

        mkdir -p $GOPATH/src/github.com/julz

        cd $GOPATH/src/github.com/julz

        git clone https://github.com/julz/pat

        cd pat

        gocart

5) Install [gcf] (https://github.com/cloudfoundry/cli)

Develop
===================================
To develop for this project, you will first need to go through the "Setting up PATs" section. This project will
be maintained through the use of standard pull requests. When issuing a pull request, make sure to include sufficient
testing through the ginkgo package (see below) to go along with any code changes. The tests should document 
functionality and provide an example of how to use it.  

1) Go through the "Setting up GO" section

2) Install [ginkgo] (http://onsi.github.io/ginkgo/#getting_ginkgo)

        go install github.com/onsi/ginkgo/ginkgo

3) Write and test your code following the ginkgo standards

4) Run all tests within the repository

	# not sure how to do this yet but we should

Run
==================================
If you wish to run PATs as a Cloud Foundry app, please refer to the section in the beginning of this guide.

To run this project, you will first need to go the "Setting up PATs" section. Afterwards, you can
change into the pat directory and run:

1) Go through the "Setting up Go" section

2) Make sure that you have targeted a cloud foundry environment from the gcf tool (# gcf login)

There are 3 options to run PATs locally:
1) Run the source code directly. 2) Compile and run an executable. 3) Run PATs with a web user interface.

- Option 1

change into the top level of this project

        cd $GOPATH/src/github.com/julz/pat

execute the command line

        go run main.go -workload gcf:push

- Option 2

Change into the main directory

        cd $GOPATH/src/github.com/julz/pat/

        go install

Run PATs executable from the command line

        pat

- Option 3

Run PATs as an HTTP server with web user interface

        go run main.go -server
        (open browser and goto http://localhost:8080)

Example calls:

	pat -h  # will output all of the command line options if installed the recommended way

	pat -concurrency=5 -iterations=5 # This will start 5 concurrent threads all pushing 1 application

	pat -concurrency=5 -iterations=1 # This will only spawn a single concurrent thread instead of the 5 you requested because you are only pushing a single application

	pat -silent # if you don't want all the fancy output use the silent flag
 
	pat -workload=gcf:push,gcf:push,... #list the gcf operations you want to run

	pat -workload=dummy    # run the tool with dummy operations (not against a CF environment)

	pat -config=config/template.yml # include a configuration template specifying any number of command line arguments. The template file provides a basic format.

	## Using the REST api:

	go run main.go -rest:target http://api.xyz.abc.net \
	  -rest:username=ibmtestuser1@us.ibm.com \
	  -rest:password=PASSWORD \
	  -workload rest:login,rest:push,rest:push,rest:push \
	  -concurrency=10 -iterations=50


Script Configure
=====================================
PATs currently accepts the ability to configure any command line argument via a configuration script. There is an example script at the root of this projects
directory called config-template.yml and it details the current operations. To run a custom yaml configuration file, provide the full path to the 
configuration file. Also, if a user so wishes they can overwrite a seeting in the script by using the command line argument.

example:
	
	# pat -config=config-template.yml
	
	# pat -config=config-template.yml -iterations=2 //set iterations to 2 even if the script has something else.
