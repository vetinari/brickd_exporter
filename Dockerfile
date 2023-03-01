FROM golang:1.19 AS build
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-extldflags '-static' -s -w" -o brickd_exporter

FROM busybox
COPY --from=build /build/brickd_exporter /usr/bin/brickd_exporter
CMD ["/usr/bin/brickd_exporter"]
