PATs (Performance Acceptance Tests)
==================================
The goal of this project is to create a super-simple load generation testing framework for quickly and easily running load against Cloud Foundry.


Setting up PATs
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
To run this project, you will first need to go the "Setting up PATs" section. Afterwards, you can
change into the pat directory and run:

1) Go through the "Setting up Go" section

2) Make sure that you have targeted a cloud foundry environment from the gcf tool (# gcf login)

Option 1

3) change into the top level of this project

	cd $GOPATH/src/github.com/julz/pat

4) execute the command line

	go run pat/main.go

Option 2

3) Change into the main directory

	cd $GOPATH/src/github.com/julz/pat/pat

	go install

4) Run PATs from the command line

	pat

5) Run PATs as an HTTP server (work in progress)

	go run pat/main.go -server # must be called in this fashion due to static file location

Example calls:

	pat -h  # will output all of the command line options if installed the recommended way

	pat -concurrency=5 -iterations=5 # This will start 5 concurrent threads all pushing 1 application

	pat -concurrency=5 -iterations=1 # This will only spawn a single concurrent thread instead of the 5 you requested because you are only pushing a single application

	pat -silent # if you don't want all the fancy output use the silent flag
 
	pat -workload=push,push,... #list the gcf operations you want to run

	pat -workload=dummy    # run the tool with dummy operations (not against a CF environment)

	pat -config=config/template.yml # include a configuration template specifying any number of command line arguments. The template file provides a basic format.


Script Configure
=====================================
PATs currently accepts the ability to configure any command line argument via a configuration script. There is an example script at the root of this projects
directory called config-template.yml and it details the current operations. To run a custom yaml configuration file, provide the full path to the 
configuration file. Also, if a user so wishes they can overwrite a seeting in the script by using the command line argument.

example:
	
	# pat -config=config-template.yml
	
	# pat -config=config-template.yml -iterations=2 //set iterations to 2 even if the script has something else.
