This is a work in progress

## Goal
Redis like caching^Wcashing  service in golang

### Example
Start the server.
(fyi listen's by default only on localhost)
```
$ go run ./cmd/main.go
```

In another shell connect to it on localhost:6379 with nc.
```
$ nc localhost 6379
set mykey somevalue34344 

get mykey
somevalue34344
del mykey
(integer) 1
get mykey
(nil)
set mykey 8

incr mykey
9
get mykey
9
```


## Ideas
TODO

## Development
### SAVE
Because redis uses fork(2) to save a copy of in memory data with a background process and given the fact golang can't fork(2) in this way I ended up implementing the in memory hash tree as copy on write with a snapshot function. SAVE or BGSAVE isn't fully implemented right now but this CoW/Snapshot behaivor should serve the same function as fork(2) upstream.

