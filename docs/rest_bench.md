# REST API benchmark

Here are some results of benchmarking GCache performance via REST API using [wrk utility](https://github.com/wg/wrk).

## Setup


* Test machine: Macbook Pro, Intel Core i5-4278U CPU @ 2.60GHz,
8Gb RAM;
* loopback networking, ulimit 7168;
* GCache settings: master, 64 shards, file dumps turned off.


## Results

### SET

Insert unique keys (worst time complexity):

```
wrk -s post_unique.lua -t10 -c30 -d60s http://localhost:8080/cache/v1/item
Running 1m test @ http://localhost:8080/cache/v1/item
  10 threads and 30 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.31ms    4.25ms  168.95ms  94.55%
    Req/Sec     1.92k     437.33  4.48k     69.70%
  1148875 requests in 1.00m, 204.92MB read
Requests/sec:  19135.48
Transfer/sec:      3.41MB
```


Update single key (best time complexity)

```
wrk -s post_update.lua -t10 -c30 -d60s http://localhost:8080/cache/v1/item
Running 1m test @ http://localhost:8080/cache/v1/item
  10 threads and 30 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.39ms  2.75ms    95.43ms   95.18%
    Req/Sec     3.21k   356.97     4.39k    78.52%
  1914267 requests in 1.00m, 337.73MB read
Requests/sec:   31895.16
Transfer/sec:       5.63MB
```

### GET

Retrieve different non-existing keys:

```
wrk -s get_var.lua -t10 -c30 -d60s http://localhost:8080/cache/v1/item
Running 1m test @ http://localhost:8080/cache/v1/item
  10 threads and 30 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.38ms   2.31ms   78.11ms   94.79%
    Req/Sec     3.02k    311.03     4.08k    75.68%
  1806133 requests in 1.00m, 284.21MB read
  Non-2xx or 3xx responses: 1806133
Requests/sec:   30091.40
Transfer/sec:       4.74MB
```

Retrieve different stored keys:

```
wrk -s get_var.lua -t10 -c30 -d60s http://localhost:8080/cache/v1/item
Running 1m test @ http://localhost:8080/cache/v1/item
  10 threads and 30 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.93ms   4.48ms   140.80ms   95.13%
    Req/Sec     2.74k    445.75     4.95k    69.57%
  1639081 requests in 1.00m, 265.68MB read
  Non-2xx or 3xx responses: 483652
Requests/sec:  27300.32
Transfer/sec:      4.43MB
```

### REMOVE

Delete unique keys from storage:

```
wrk -s remove_var.lua -t10 -c30 -d60s http://localhost:8080/cache/v1/item
Running 1m test @ http://localhost:8080/cache/v1/item
  10 threads and 30 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.34ms   2.58ms   97.11ms   95.16%
    Req/Sec     3.22k    336.55    5.47k    76.98%
  1923526 requests in 1.00m, 117.40MB read
Requests/sec:  32044.70
Transfer/sec:      1.96MB
```


## Test scripts

Lua scripts used for benchmarking:

**get_var.lua**

```lua
keyCount = 0

request = function()
    wrk.body = string.format("{\"key\": \"testkey%d\"}", keyCount)
    wrk.headers["Content-Type"] = "application/json"
    keyCount = keyCount + 1
    return wrk.format("GET",nil)
end

```

**post_update.lua**

```lua
wrk.method = "POST"
wrk.body   = "{\"key\": \"sometestkey\", \"value\": \"some test string here\", \"ttl\": 60}"
wrk.headers["Content-Type"] = "application/json"
```

**post_unique.lua**

```lua
keyCount = 0

request = function()
    wrk.body = string.format("{\"key\": \"testkey%d\", \"value\": \"some test string here\", \"ttl\": 600}", keyCount)
    wrk.headers["Content-Type"] = "application/json"
    keyCount = keyCount + 1
    return wrk.format("POST",nil)
end

```

**remove_var.lua**

```lua
keyCount = 0

request = function()
    wrk.body   = string.format("{\"key\": \"testkey%d\"}", keyCount)
    wrk.headers["Content-Type"] = "application/json"
    keyCount = keyCount + 1
    return wrk.format("DELETE",nil)
end
```
