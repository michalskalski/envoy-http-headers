## Envoy http response headers modification

This repository contains simple scenario consisting of:
* [envoy proxy](envoy) conifigured with external_authz (listen on port 8081)
* simple [grpc servise](authz_server) acting as external authorization service (listen on port 5001)
* [http backend](backend_server) generating reponses (listen on port 9091)

This setup tests behavior of [header append actions](https://www.envoyproxy.io/docs/envoy/v1.29.1/api-v3/config/core/v3/base.proto#envoy-v3-api-enum-config-core-v3-headervalueoption-headerappendaction) which is done in external authorization service to the replay to client. `ADD_IF_ABSENT` and `APPEND_IF_EXISTS_OR_ADD` instead of only adding header when it is missing from backend response or append to existing value, it always set header defined in `ResponseHeadersToAdd` list.

Issue reported here https://github.com/envoyproxy/envoy/issues/32657 .

### How to reproduce

Clone this repository and run:
```sh
docker-compose up -d --force-recreate --build
```

it will bring up all components. [Backend](backend_server) is configured to include `test-header: original` header in response when `GET /`, and not add it when `GET /no`. Authorization server looks into request headers, to find `action` header which can take two values:
* `add-if-absent`
* `append-if-exist`

and based on this value will chose coresponding append action for `test-header`.

### Expected behavior:

```sh
curl -v -H "action: add-if-absent" localhost:8081
```

response should contain `test-header: original`

```sh
curl -v -H "action: append-if-exist" localhost:8081
```

response should contain `test-header: original append-if-exist`

### Current behavior

#### ADD_IF_ABSENT
```sh
curl -v -H "action: add-if-absent" localhost:8081
*   Trying [::1]:8081...
* Connected to localhost (::1) port 8081
> GET / HTTP/1.1
> Host: localhost:8081
> User-Agent: curl/8.4.0
> Accept: */*
> action: add-if-absent
>
< HTTP/1.1 200 OK
< backend: yes
< date: Fri, 01 Mar 2024 07:42:00 GMT
< content-length: 26
< content-type: text/plain; charset=utf-8
< x-envoy-upstream-service-time: 7
< test-header: add-if-absent
< server: envoy
<
* Connection #0 to host localhost left intact
responded with test-header
```

envoy logs:
```
[2024-03-01 07:42:00.202][15][trace][router] [source/common/router/upstream_request.cc:261] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] upstream response headers:
':status', '200'
'backend', 'yes'
'test-header', 'original'
'date', 'Fri, 01 Mar 2024 07:42:00 GMT'
'content-length', '26'
'content-type', 'text/plain; charset=utf-8'

[2024-03-01 07:42:00.202][15][debug][router] [source/common/router/router.cc:1506] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] upstream headers complete: end_stream=false
[2024-03-01 07:42:00.202][15][trace][misc] [source/common/event/scaled_range_timer_manager_impl.cc:60] enableTimer called on 0x16cd3f260a80 for 300000ms, min is 300000ms
[2024-03-01 07:42:00.202][15][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:221] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] ext_authz filter has 0 response header(s) to add and 1 response header(s) to set to the encoded response:
[2024-03-01 07:42:00.202][15][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:233] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] ext_authz filter set header(s) to the encoded response:
[2024-03-01 07:42:00.202][15][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:235] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] 'test-header':'add-if-absent'
[2024-03-01 07:42:00.202][15][trace][http] [source/common/http/filter_manager.cc:1233] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] encode headers called: filter=envoy.filters.http.ext_authz status=0
[2024-03-01 07:42:00.202][15][debug][http] [source/common/http/conn_manager_impl.cc:1869] [Tags: "ConnectionId":"1","StreamId":"2892000782072673515"] encoding headers via codec (end_stream=false):
':status', '200'
'backend', 'yes'
'date', 'Fri, 01 Mar 2024 07:42:00 GMT'
'content-length', '26'
'content-type', 'text/plain; charset=utf-8'
'x-envoy-upstream-service-time', '7'
'test-header', 'add-if-absent'
'server', 'envoy'
```

#### APPEND_IF_EXISTS_OR_ADD
```sh
curl -v -H "action: append-if-exist" localhost:8081
*   Trying [::1]:8081...
* Connected to localhost (::1) port 8081
> GET / HTTP/1.1
> Host: localhost:8081
> User-Agent: curl/8.4.0
> Accept: */*
> action: append-if-exist
>
< HTTP/1.1 200 OK
< backend: yes
< date: Fri, 01 Mar 2024 07:44:23 GMT
< content-length: 26
< content-type: text/plain; charset=utf-8
< x-envoy-upstream-service-time: 1
< test-header: append-if-exist
< server: envoy
<
* Connection #0 to host localhost left intact
responded with test-header% 
```

```
[2024-03-01 07:44:23.495][19][trace][router] [source/common/router/upstream_request.cc:261] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] upstream response headers:
':status', '200'
'backend', 'yes'
'test-header', 'original'
'date', 'Fri, 01 Mar 2024 07:44:23 GMT'
'content-length', '26'
'content-type', 'text/plain; charset=utf-8'

[2024-03-01 07:44:23.495][19][debug][router] [source/common/router/router.cc:1506] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] upstream headers complete: end_stream=false
[2024-03-01 07:44:23.495][19][trace][misc] [source/common/event/scaled_range_timer_manager_impl.cc:60] enableTimer called on 0x16cd3f260e80 for 300000ms, min is 300000ms
[2024-03-01 07:44:23.496][19][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:221] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] ext_authz filter has 0 response header(s) to add and 1 response header(s) to set to the encoded response:
[2024-03-01 07:44:23.496][19][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:233] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] ext_authz filter set header(s) to the encoded response:
[2024-03-01 07:44:23.496][19][trace][ext_authz] [source/extensions/filters/http/ext_authz/ext_authz.cc:235] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] 'test-header':'append-if-exist'
[2024-03-01 07:44:23.496][19][trace][http] [source/common/http/filter_manager.cc:1233] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] encode headers called: filter=envoy.filters.http.ext_authz status=0
[2024-03-01 07:44:23.496][19][debug][http] [source/common/http/conn_manager_impl.cc:1869] [Tags: "ConnectionId":"4","StreamId":"3447697829219430721"] encoding headers via codec (end_stream=false):
':status', '200'
'backend', 'yes'
'date', 'Fri, 01 Mar 2024 07:44:23 GMT'
'content-length', '26'
'content-type', 'text/plain; charset=utf-8'
'x-envoy-upstream-service-time', '1'
'test-header', 'append-if-exist'
'server', 'envoy'
```
