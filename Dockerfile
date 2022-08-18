############################
# STEP 1 build executable binary
############################
FROM golang as builder

WORKDIR $GOPATH/srv/arsiesys/nomad-node-state-exporter/
COPY . .

RUN go mod vendor
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/nomad-node-state-exporter -mod vendor main.go

############################
# STEP 2 build a small image
############################
FROM scratch

ENV GIN_MODE=release
WORKDIR /app/
# Import from builder.
COPY --from=builder /go/bin/nomad-node-state-exporter /app/nomad-node-state-exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app/nomad-node-state-exporter"]
EXPOSE 9827
