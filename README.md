This project is still under development.

# Gimulator

The main idea of this project is holding AI competitions where one or more agents want to communicate with the environment or other agents. Gimulator is a Key-Value store with **Watch**, **Get**, **Set**, **Delete**, and **Find** functionality, Also it can Authenticates and Authorizes incoming requests. It is at the center of any system and is responsible for transmitting and storing messages. Every message has a unique key and some value. Value can contain any type of data like JSON, Yaml, or string.

# Getting Started

You can get the project using this command:

```bash
$ go get github.com/Gimulator/Gimulator
```

After that, Start the project using these commands:

```bash
$ cd $GOPATH/src/github.com/Gimulator/Gimulator/example
$ go run main.go ./roles.yaml
```

Now Gimulator listens and serves on `http://localhost:5000/`

# Description

In this section, Gimulator is briefly described.

## Object

Object is the core entity of Gimulator. Object has this structure:

```
object:
    key:
        type        string
        name        string
        namespace   string
    value           interface{} #It means any type of data, like JSON or Yaml.
```

Key is the identifier of objects. Gimulator can find objects based on keys, It just filters all keys to find the key which is requested. You can define your language using keys and the different components of your system can speak to each other with this language, through connecting to Gimulator. 

## Operations

 Gimulator supports five operations:

 1. Get: To get an object with a specific key.
 2. Set: To set an object with a specific key.
 3. Find: To get a list of objects which match with a specific key.
 4. Delete: To delete an object with a specific key.
 5. Watch: For setting a new watcher for a specific key to be aware of changes of that object.

## Components

Gimulator contains four main packages:

1. storage: This package has to store all incoming objects. it already saves objects in a map data structure in RAM.
2. simulator: It is a middleware package between **storage** and **api** packages. This package has to transmit incoming requests from api to storage package, and if there is a set operation on an object, it should push the new object to the clients who watch on this object.
3. auth: This package has to authenticate new clients and authorize every request from clients based on a config file.
4. api: This package handles incoming HTTP requests. Every request's method is "POST". List of endpoints:

    * `/get`
    * `/set`
    * `/find`
    * `/watch`
    * `/delete`
    * `/register`
    * `/socket`

At first, you should send a "POST" request with credentials in its body to the `/register` endpoint. the api package takes this request and checks the credentials with the auth package, and if auth returns OK, api send you a token and you should save the token for sending future requests.
The `/socket` endpoint opens a web-socket connection to send any changes you want to watch.

# Contribute

We've set up a separate document for our [contribution guidelines](https://github.com/Gimulator/Gimulator/blob/readme/CONTRIBUTING.md).