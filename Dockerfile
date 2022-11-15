FROM golang:1.18-alpine as builder
COPY . /go/src/eci.io/eci-profile
WORKDIR /go/src/eci.io/eci-profile
RUN apk add make && make && cp ./bin/eci-profile /eci-profile

FROM alpine:3.16
COPY --from=builder /eci-profile /usr/bin/eci-profile
ENTRYPOINT [ "/usr/bin/eci-profile" ]