# BlotWeb

BlotWeb is a debugging tool for working with [etcd-bbolt](https://github.com/etcd-io/bbolt). 
It is inspired by [evnix/boltdbweb](https://github.com/evnix/boltdbweb) and supports nested buckets.

BlotWeb provides web interface for interacting with your bolt database.

## Web Interface

Features:
- List all top-level buckets in the database
- List all keys and values in a specified nested bucket
- Keys with links indicate nested buckets, which can be clicked to view the contents of the nested bucket


### List top-level buckets

`GET /buckets/`

This endpoint lists all top-level buckets in the database.

### List keys and values in a nested bucket

`GET /buckets/<bucket1>/<bucket2>/...`

This endpoint lists all key-value pairs in the specified nested bucket. 
The URL path should include the names of the nested buckets, separated by forward slashes (`/`).


## Installation

To install BlotWeb, use the following command:
```
go install github.com/flymop/blotweb@latest
```
This will start the BlotWeb server on the default port (9092). 
You can access the web interface by opening a web browser and navigating to `http://localhost:9092/buckets`.

and run:
```
blotweb -db <your db file>
```


## Other
By default, BlotWeb assumes that all key-value pairs in the database are in string format. 
However, you can modify this behavior by changing the `toString` function in the code to implement your own conversion logic.


