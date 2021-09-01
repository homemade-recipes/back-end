FROM golang:1.16 as builder

COPY . /root/

RUN cd /root && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch

COPY --from=builder /root/service /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 8080
EXPOSE 27017

CMD ["/service"]