---------------------------------------

restatic -- thinobject web viewer using go net/http

adapted from relogHQ/restatic

---------------------------------------

#  What is thinobject?

Thinobject is a system using ordinary files, directories, symlinks, and exectutable
methods to effect an object-oriented interface to the filesystem, by adopting several
conventions:

    1. an object is a directory

    1. a non-resolving symlink named ^ (caret) identifies the object type

    1. types resolve as directories under environment variable $TOBLIB 

    1. a thinobject type is a directory containing executable programs or scripts

    1. methods in a thinobject type directory can be run by an object vi the bin/tob script

    1. tob hooks bash's command_not_found_handle() to run methods as: 'object.method arg...'

    1. tob changes the working directory to the object before the method is run

    1. a text file starting with @ in the name is treated as a list of lines

    1. a text file starting with % in the name is treated as a map of 'key value...' lines

    1. symlinks are used as string variables, 'symvars', with = as prefix to the value

#  What is Restatic?

Restatic is a simple HTTP server that serves a local directory over HTTP. It is written in
[Go](https://golang.org/), using go's net/http and other standard packages.

The Restatic server provide a web interface for the directory it is started in, showing the
name, size, and creation date of each file.

##  Using Restatic

```
$ ./restatic -p 4001 -d .

INFO[0000] server listening on :4001  
INFO[0000] =========================
```

 4. Browsing the URL on your favourite browser will load a webpage like this

![screen-01](https://user-images.githubusercontent.com/4745789/135251623-f8ea8024-75b7-4150-a869-26135212822d.PNG)

##  Developing Restatic

If you are a developer and want to modify restatic, you will have first to set up a dev environment, and it has the following pre-requisites

- [Go 1.17.1](https://golang.org/)

Once you have set up all the pre-requisites, following the steps to start your development server.

- Clone the repository https://github.com/relogHQ/restatic
- Start the server `go run cmd/restatic/main.go`

Once you start the server, it will download all the necessary packages and listen to the configured port. The default port is 5030.

##  Linting

Maintaining coding standards is extremely critical, and restatic follows the standard [Gofmt](https://pkg.go.dev/cmd/gofmt) to reformat the code. It is customary to fire the following command before you commit.

```
make lint
```

##  Contribution Guidelines

The Code Contribution Guidelines are published at [CONTRIBUTING.md](https://github.com/relogHQ/restatic/blob/master/CONTRIBUTING.md); please read them before you start making any changes. This would allow us to have a consistent standard of coding practices and developer experience.

##  Relog Umbrella
<div align="center">
<br />
<img  width="240"  src="https://user-images.githubusercontent.com/4745789/133601178-711aa4eb-f836-4e93-a554-22006648f75f.png" align="center"  alt="relog logo" />
<br />
<br />
</div>

[Relog](https://relog.in) is an initiative that aims to transform engineering education and provide high-quality engineering courses, projects, and resources to the community. To better understand all the common systems, we aim to build our own replicated utilities, for example, a load balancer, static file server, API rate limiter, etc. All the projects fall under [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0), and you can find their source code at [github.com/relogHQ](https://github.com/relogHQ).

##  License
Restatic is under [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)
