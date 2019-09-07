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
````
$ nc localhost 6379
add 12 123494
OK
get 12
123494
remove 12
123494
get 12
-1
add 11 123484
OK
get 11
123484
evict
OK
get 11
-1
```


## Ideas
TODO

## Development
TODO

