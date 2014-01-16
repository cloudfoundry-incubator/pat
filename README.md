PATs (Performance Acceptance Tests)
==================================
The goal of this project is to create a load generation testing framework.


Setting up PATs
==================================
To setup this project, a number of requires need to be met for GO.

1) Ensure that GO (64bit version) has been installed on the system.

2) Setup the GOPATH

        # export GOPATH=~/go (or any other workspace repository for all your go code)

        # export PATH=$GOPATH/bin:$PATH

3) Install [gocart] (https://github.com/vito/gocart)

        # go get github.com/vito/gocart

4) Clone the project to the correct location

        # mkdir -p $GOPATH/src/github.com/julz

        # cd $GOPATH/src/github.com/julz

        # git clone "This project" and change into the new pat directory

        # gocart

Develop
===================================
To develop for this project, you will first need to go through the "Setting up PATs" section. This project will
be maintained through the use of standard pull requests. When issuing a pull request, make sure to include sufficient
testing through the ginkgo package (see below) to go along with any code changes. The tests should document 
functionality and provide an example of how to use it.  

1) Go through the "Setting up GO" section

2) Install [ginkgo] (http://onsi.github.io/ginkgo/#getting_ginkgo)

        # go install github.com/onsi/ginkgo/ginkgo

3) Write and test your code following the ginkgo standards

4) Run all tests within the repository

	# not sure how to do this yet but we should

Run
==================================
To run this project, you will first need to go the "Setting up PATs" section. Afterwards, you can
change into the pat directory and run:

1) Go through the "Setting up Go" section

2) change into the top level of this project

	# cd $GOPATH/src/github.com/julz/pat

3) execute the command line

	# go run pat/main.go

OR install the binary (recommended)

2) Change into the main directory

	# cd $GOPATH/src/github.com/julz/pat/pat

	# go install

3) Run PATs from the command line

	# pat

4) Run PATs as an HTTP server (work in progress)

	# pat -server

Example calls:

	# pat -h [will output all of the command line options if installed the recommended way]

	# pat -concurrency=5 -pushes=5 [This will start 5 concurrent threads all pushing 1 application]

	# pat -concurrency=5 -pushes=1 [This will only spawn a single concurrent thread instead of the 5 you requested because you are only pushing a single application]

	# pat -silent [if you don't want all the fancy output use the silent flag]
 

