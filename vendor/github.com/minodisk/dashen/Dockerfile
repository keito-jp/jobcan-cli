FROM golang:1.7.5
WORKDIR /go/src/github.com/minodisk/dashen

RUN apt-get -y update && \
    apt-get -y install \
      libpcap0.8-dev
RUN go get -u \
      github.com/google/gopacket \
      github.com/pkg/errors

COPY . .

CMD go test -v -race ./...
