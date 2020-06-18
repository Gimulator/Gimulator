# Gimulator

:warning: **_This project is still under development._**

The main purpose of this project is to hold AI competitions where one or more agents need to communicate with the environment or other agents. Gimulator is a Key-Value store with an API consisting of **Watch**, **Get**, **Set**, **Delete**, and **Find** endpoints to work with. It can also authenticate and authorize incoming requests. It is responsible for transmitting and storing messages. Every message has a unique key and a value. Value can contain any type of data like JSON, Yaml, or string.

## Table of Contents

- [Getting Started](#getting-started)
- [Description](#description)
  - [Object](#object)
  - [Operations](#operations)
  - [Components](#components)
- [Contributing](#contributing)

## Getting Started

### 1. Get the project using this command:

```bash
$ go get github.com/Gimulator/Gimulator
```
or you can get it with git command:
```
$ git clone https://github.com/Gimulator/Gimulator.git
```

### 2. Then, start it using these commands:

```bash
$ make build
$ GIMULATOR_HOST="localhost:3030" GIMULATOR_ROLES_FILE_PATH="{/path/to/roles-file.yaml}" make exec
```
or you can run it directly with this command:
```
$ GIMULATOR_HOST="localhost:3030" GIMULATOR_ROLES_FILE_PATH="{/path/to/roles-file.yaml}" make run
```

Gimulator is up and running! You can access the API on `localhost:{port}`.

## Description

In this section, Gimulator is briefly described.

### Object

Object is the core entity of Gimulator. Object has this structure:

```
object:
    key:
        type        string
        name        string
        namespace   string
    value           interface{} # It means any type of data, like JSON or Yaml.
    meta:
        creation-time: string # Example: 2020-06-18 18:57:41.975664865 # Time of creating this object in storage
        owner:         string # Who creates this object
        method:        string # It is 'set' or 'delete'
```

Key is the identifier of objects. Gimulator can find objects based on keys. It just filters all keys to find the key which is requested. You can define your language using keys and the different components of your system can speak to each other with this language, through connecting to Gimulator.
Meta is filled by Gimulator and if you fill it client-side, Gimualtor overwrites it.
### Operations

Gimulator supports five operations:

 1. Get: To get an object with a specific key().
 2. Set: To set an object with a specific key.
 3. Find: To get a list of objects which match with a partial key.
 4. Delete: To delete an object with a specific key.
 5. Watch: To set a new watcher for a specific key to be notified of changes of the objects filtered by that key.

When you want to use `set`, `get`, or `delete`, you should fill all the key's entities(Name, Namespace, Type).

### Components

Gimulator contains four main packages:

1. storage: This package stores all incoming objects. It currently saves objects in a map data structure in memory.
2. simulator: It is a middleware package between **storage** and **api** packages. This package has to transmit incoming requests from api to storage package, and if there is a set operation on an object, it should push the new object to the clients who watch on this object.
3. auth: This package authenticates new clients and authorizes every request from clients based on a config file.
4. api: This package handles incoming HTTP requests. All the endpoint's method are "POST". List of endpoints:

    * `/get`
    * `/set`
    * `/find`
    * `/watch`
    * `/delete`
    * `/register`
    * `/socket`

At first, you should send a "POST" request with credentials in its body to the `/register` endpoint. the api package takes this request and checks the credentials with the auth package, and if auth returns OK, API returns a token and you should save the token for sending future requests.
The `/socket` endpoint opens a Websocket connection to send any changes you want to watch.

## Contributing

We've set up a separate document for our [contribution guidelines](https://github.com/Gimulator/Gimulator/blob/readme/CONTRIBUTING.md).
