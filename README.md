# GCache - cache server in Go

## Description
Yet another Redis-like in-memory cache server written in Go.
Could be used for saving sessions, counters and other types of information with limited lifetime and very fast access. GCache supports easy scaling for data retrieval operations with master-slave replication.


## Features

* Key-value storage with data expiration.
* Store strings, numbers, booleans, arrays and dictionaries with string keys.
* Basic CRUD operations.
* Additional data retrieval operations on complex values: arrays and dictionaries.
* Optional persistence with periodic saving of snapshots on the disk.
* Restore cache state from file on start.
* Master-slave replication.
* REST API.
* Native client library written in Go.


## Installation

For building cache server Go 1.8+ is required.

```
go get github.com/dgtony/gcache
```

After installation GCache could be started with defined configuration:

```
gcache -c <path_to_config_file.toml>
```


## Usage

### Basic operations
GCache currently supports following basic operations:

* GET - retrieve stored value by key;
* SET - store new value with given key and TTL in seconds;
* REMOVE - delete entire stored value;
* KEYS - retrieve stored keys.

Operation SET performed on existing key will update its value and TTL.

There are two variations of KEYS command: using plain version will return all keys stored in cache, while adding mask parameter enforce server to return only matching keys. Glob pattern matching rules used to create mask, similar to Redis.

**Note:** as usual, KEYS could be a very expensive kind of operation, use it with care!


### Sub-element access

Internal elements of complex data types, such as arrays and dictionaries, could be efficiently accessed with additional methods:

* GETSUBINDEX - return element of value array with given sub-index
* GETSUBKEY - return element of value dictionary with given sub-key


**Note:** subindexing in value array starts from 1, element index 0 will return entire array!


### Data persistence

In order to add some persistence data stored in cache memory could be periodically dumped in file. As the node starts it may use such snapshot file to restore data from the previous session. Please mind, that after cache re-establishment all outdated keys will be removed.

Dumping and restoring options could be set in node's configuration file.

**Note**: slave nodes (see below) do not restore its cache from file, but only from the master-node.


### Replication

Each cache node could be started in one of the following modes:

* *standalone*
* *master*
* *slave*

Standalone node works as a single server and makes no communication with other nodes.

For the purposes of scaling data retrieval process GCache could be horizontally sharded in a cluster. Cluster consists of a single *master-node* and several *slave-nodes*. Data modifying operations, such as GET/REMOVE are allowed only on master node, while values could be retrieved from slaves, as well. System is based on eventual consistency model, where slaves periodically update its cache from master node.
One can adjust data inconsistency window with `dump_update_period` parameter in node configuration file.


### REST API

GCache provides REST API as a standard server access interface. Swagger-powered API specification could be found in file `docs/rest_api.html`.


### Native clients

At the moment the only existing native client is *gclient* – thin library written in Go. More information about library and usage examples could be found in the project [repository](https://github.com/dgtony/gclient).


### TODO

Project just started, and there are a lot of things to be done. Among others:

 * client authorization;
 * efficient sub-element modification in complex values;
 * differential cache updates;
 * cache snapshotting triggered by data modification;
 * custom binary protocol for native libraries;
 * etc.

Any help is welcome :)
