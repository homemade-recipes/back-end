FROM golang:1.16 as builder

COPY . /root/

RUN cd /root && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch

COPY --from=builder /root/back-end /
COPY --from=builder /root/db.json /

CMD ["/back-end"]