# promtail-client

Promtail client library. Promtail is an agent for [Loki](https://github.com/grafana/loki) logging system.

This library supports both JSON and Protobuf APIs.

Usage example:

```
cd examples/client-example
go build

# make sure source-name is unique for every application instance
# otherwise promtail will reject logs with error:
# "entry out of order for stream"
./client-example proto source-name job-name
```
