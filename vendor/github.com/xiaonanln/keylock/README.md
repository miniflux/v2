# keylock
Golang utility class KeyLock: lock by string key, so as to avoid giant lock

## Testing

Since this utility deals with concurrency so much, it is important to run the tests with the `-race` flag:

    $ go test -race
