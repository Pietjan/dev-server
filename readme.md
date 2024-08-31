# dev-server

rebuilds and runs a command on file changes in the current working directory.

## Install
``` sh
go install github.com/pietjan/dev-server@latest
```

## Run
``` sh
dev-server --build.cmd "make build" --build.bin "/tmp/bin/my-program"
```

full example
``` sh
dev-server \
--build.cmd "make build" \
--build.bin "/tmp/bin/my-program" \
--watcher.exclude "\/\.(\w+)$" \
--watcher.exclude "^bin/" \
--watcher.interval 400ms \
--proxy.port 42069 \
--proxy.target 8080 \
--wait.for localhost:5432 \
--wait.for localhost:3306 \
--debug
```
