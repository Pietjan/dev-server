# dev-server

rebuilds and runs a golang application on file changes in the current working directory.

``` sh
go install github.com/pietjan/dev-server@latest
```

## Config
The dev-server config can specified in either json or yml/yaml

.dev-server.yml / .dev-server.yaml
``` yaml 
interval: 500 
exclude: 
  - ^bin/
  - ^\.
  - \/\.(\w+)$]
build: ./
target: ./bin/__dev-server_target
server: 42069
proxy: 8080
```

.dev-server.json
``` json 
{
  "interval": 500,
  "exclude": [
    "^bin/",
    "^\\.",
    "\\/\\.(\\w+)$"
  ],
  "build": "./",
  "target": "./bin/__dev-server_target",
  "server": 42069,
  "proxy": 8080
}
```