# Nomad node state exporter

Prometheus exporter for Nomad nodes

This exporter listen on port `9827` and the endpoint is `/metrics`

```
# HELP nomad_node_busy_status The busy status of a nomad node 0:IDLE 1:BUSY
# TYPE nomad_node_busy_status gauge
nomad_node_busy_status{Datacenter="DC1",computerName="BLDXXXX",nodeClass="Staging"} 0
# HELP nomad_node_exporter_failure The number of failure to get/parse api data since startup
# TYPE nomad_node_exporter_failure counter
nomad_node_exporter_failure 0
# HELP nomad_node_maintenance_status The maintenance status of a nomad node 0:ONLINE 1:MAINTENANCE 2:OFFLINE
# TYPE nomad_node_maintenance_status gauge
nomad_node_maintenance_status{Datacenter="DC1",computerName="BLDXXXX",nodeClass="Staging"} 0
```

Available options:
```
Generate prometheus metrics for nomad nodes states

Usage:
  nomad-node-state-exporter [flags]

Flags:
  -a, --address string            address of the nomad server api (default "https://my-nomad-server:4646")
      --ca string                 Trusting CA certificate for TLS auth (default "/nomad-pki/nomad-ca.pem")
      --cert string               Certificate used for TLS auth (default "/nomad-pki/cli.pem")
      --disable-authentication    disable authentication
      --fetch-interval duration   fetch-interval in seconds (default 30s)
  -f, --filter string             Nomad format expression filter for allocations endpoint. example: Name contains "jenkins"
  -h, --help                      help for nomad-node-state-exporter
      --key string                Certificate KEY used for TLS auth (default "/nomad-pki/cli-key.pem")
      --port int                  port to listen on (default 9827)
```
