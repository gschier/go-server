FROM golang:1.21 as app-builder
WORKDIR /go/src/app
COPY . .
RUN echo "Cache break counter: 7"
# Static build required so that we can safely copy the binary over.
# `-tags timetzdata` embeds zone info from the "time/tzdata" package.
RUN CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' -tags timetzdata ./...

FROM scratch
# the test program:
COPY --from=app-builder /go/bin/web /go-server-web
COPY --from=app-builder /go/bin/worker /go-server-worker
# the tls certificates:
# NB: this pulls directly from the upstream image, which already has ca-certificates:
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/go-server-web"]
