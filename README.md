[![portal logo](http://nano-assets.gopagoda.io/readme-headers/portal.png)](http://nanobox.io/open-source#portal)  
[![Build Status](https://travis-ci.org/nanopack/portal.svg)](https://travis-ci.org/nanopack/portal)
[![GoDoc](https://godoc.org/github.com/nanopack/portal?status.svg)](https://godoc.org/github.com/nanopack/portal)

Portal is an api-driven, in-kernel layer 2/3 load balancer.

## Status
Complete/Experimental

## Usage:

### As a CLI
Simply run `portal <COMMAND>`

`portal` or `portal -h` will show usage and a list of commands:

```
portal - load balancer/proxy

Usage:
  portal [flags]
  portal [command]

Available Commands:
  add-service    Add service
  remove-service Remove service
  show-service   Show service
  show-services  Show all services
  set-services   Set service list
  set-service    Set service
  add-server     Add server to a service
  remove-server  Remove server from a service
  show-server    Show server on a service
  show-servers   Show all servers on a service
  set-servers    Set server list on a service
  add-route      Add route
  set-routes     Set route list
  show-routes    Show all routes
  remove-route   Remove route
  add-cert       Add cert
  set-certs      Set cert list
  show-certs     Show all certs
  remove-cert    Remove cert

Flags:
  -C, --api-cert="": SSL cert for the api
  -H, --api-host="127.0.0.1": Listen address for the API
  -k, --api-key="": SSL key for the api
  -p, --api-key-password="": Password for the SSL key
  -P, --api-port="8443": Listen address for the API
  -t, --api-token="": Token for API Access
  -r, --cluster-connection="none://": Cluster connection string (redis://127.0.0.1:6379)
  -T, --cluster-token="": Cluster security token
  -c, --conf="": Configuration file to load
  -d, --db-connection="scribble:///var/db/portal": Database connection string
  -i, --insecure[=false]: Disable tls key checking (client) and listen on http (server)
  -L, --log-file="": Log file to write to
  -l, --log-level="INFO": Log level to output
  -s, --server[=false]: Run in server mode

Use "portal [command] --help" for more information about a command.
```

For usage examples, see [Api](api/README.md) and/or [Cli](commands/README.md) readme  

### As a Server
To start portal as a server run:

`portal --server`

An optional config file can also be passed on startup:

`portal -c /path/to/config.json`

>config.json
>```json
{
  "api-token": "",
  "api-host": "127.0.0.1",
  "api-port": 8443,
  "api-key": "",
  "api-cert": "",
  "api-key-password": "",
  "db-connection": "scribble:///var/db/portal",
  "cluster-connection": "none://",
  "cluster-token": "",
  "insecure": false,
  "log-level": "INFO",
  "log-file": "",
  "server": true
}
```

## API:

| Route | Description | payload | output |
| --- | --- | --- | --- |
| **Get** /services | List all services | nil | json array of service objects |
| **Post** /services | Add a service | json service object | json service object |
| **Put** /services | Reset the list of services | json array of service objects | json array of service objects |
| **Put** /services/:service_id | Reset the specified service | nil | json service object |
| **Get** /services/:service_id | Get information about a service | nil | json service object |
| **Delete** /services/:service_id | Delete a service | nil | success message or an error |
| **Get** /services/:service_id/servers | List all servers on a service | nil | json array of server objects |
| **Post** /services/:service_id/servers | Add new server to a service | json server object | json server object |
| **Put** /services/:service_id/servers | Reset the list of servers on a service | json array of server objects | json array of server objects |
| **Get** /services/:service_id/servers/:server_id | Get information about a server on a service | nil | json server object |
| **Delete** /services/:service_id/servers/:server_id | Delete a server from a service | nil | success message or an error |
| **Delete** /routes | Delete a route | subdomain, domain, and path (json or query) | success message or an error |
| **Get** /routes | List all routes | nil | json array of route objects |
| **Post** /routes | Add new route | json route object | json route object |
| **Put** /routes | Reset the list of routes | json array of route objects | json array of route objects |
| **Delete** /certs | Delete a cert | json cert object | success message or an error |
| **Get** /certs | List all certs | nil | json array of cert objects |
| **Post** /certs | Add new cert | json cert object | json cert object |
| **Put** /certs | Reset the list of certs | json array of cert objects | json array of route objects |

- **service_id** is a formatted combination of service info: type-host-port. (tcp-127_0_0_3-80)  
- **server_id** is a formatted combination of server info: host-port. (192_0_0_3-8080)  

For examples, see [the api's readme](api/README.md)  

## Data types:
### Service:
json:
```json
{
  "host": "127.0.0.1",
  "port": 1234,
  "type": "tcp",
  "scheduler": "wlc",
  "persistence": 300,
  "netmask": "255.255.255.0",
  "servers": []
}
```

json with a server:
```json
{
  "host": "127.0.0.1",
  "port": 8080,
  "type": "tcp",
  "scheduler": "wlc",
  "persistence": 300,
  "netmask": "",
  "servers": [
    {
      "host": "172.28.128.4",
      "port": 8081,
      "forwarder": "m",
      "weight": 1,
      "UpperThreshold": 0,
      "LowerThreshold": 0
    }
  ]
}
```

Fields:
- **host**: IP of the host the service is bound to.
- **port**: Port that the service listens to.
- **type**: Type of service. Either tcp or udp.
- **scheduler**: How to pick downstream server. On of the following: rr, wrr, lc, wlc, lblc, lblcr, dh, sh, sed, nq
- **persistence**: Timeout for keeping requests from the same client going to the same server
- **netmask**: How to group clients with persistence to servers
- **servers**: Array of server objects associated to the service

### Server:
json:
```json
{
  "host": "127.0.0.1",
  "port": 1234,
  "forwarder": "m",
  "weight": 1,
  "upper_threshold": 0,
  "lower_threshold": 0
}
```

Fields:
- **host**: IP of the host the service is bound to.
- **port**: Port that the service listens to.
- **forwarder**: Method to use to forward traffic to this server. One of the following: g (gatewaying), i (ipip), m (masquerading)
- **weight**: Weight to perfer this server. Set to 0 if no traffic should go to this server.
- **upper_threshold**: Stop sending connections to this server when this number is reached. 0 is no limit.
- **lower_threshold**: Restart sending connections when drains down to this number. 0 is not set.

### Route:
json:
```json
{
  "subdomain": "admin",
  "domain": "test.com",
  "path": "/admin*",
  "targets": ["http://127.0.0.1:8080/app1","http://127.0.0.2"],
  "fwdpath": "/",
  "page": ""
}
```

Fields:
 - **subdomain**: Subdomain to match on
 - **domain**: Domain to match on
 - **path**: Path to match on
 - **targets**: URIs of servers
 - **fwdpath**: Path to forward to targets (combined with target path)
 - **page**: Page to serve instead of routing to targets

### Certificate:
json:
```json
{
  "key": "-----BEGIN PRIVATE KEY-----\nMII.../J8\n-----END PRIVATE KEY-----",
  "cert": "-----BEGIN CERTIFICATE-----\nMII...aI=\n-----END CERTIFICATE-----"
}
```

Fields:
 - **key**: Pem style key
 - **cert**: Pem style certificate

### Error:
json:
```json
{
  "error": "exit status 2: unexpected argument"
}
```

Fields:
 - **error**: Error message

### Message:
json:
```json
{
  "msg": "Success"
}
```

Fields:
 - **msg**: Success message

[![portal logo](http://nano-assets.gopagoda.io/open-src/nanobox-open-src.png)](http://nanobox.io/open-source)
