# nomad-grpc-firehose-nozzle-poc
A proof-of-concept using Nomad's beta gRPC-based event stream endpoint

# Using

## Clone this repo

```
git clone git@github.com:krishicks/nomad-grpc-firehose-nozzle-poc
```

## Build and run Nomad
Using this requires building Nomad off of an as-yet unmerged branch, event/proto-service.

```
git clone git@github.com:hashicorp/nomad
cd nomad
git checkout event/proto-service
make deps
make dev
cd ../nomad-grpc-firehose-nozzle-poc
../nomad/bin/nomad agent -dev -config server.hcl
# wait for Nomad to start
# ...
#     2020-11-20T09:48:49.157-0800 [DEBUG] client: state changed, updating node and re-registering
#     2020-11-20T09:48:49.158-0800 [INFO]  client: node registration complete
# ready
```

## Run the nozzle

```
# in a new terminal
cd nomad-grpc-firehose-nozzle-poc
go run main.go
```

## Submit a Nomad job

```
# in a new terminal
./nomad/bin/nomad job run ./nomad-grpc-firehose-nozzle-poc/example.nomad
# see new events in the nozzle terminal
```
