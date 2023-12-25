FROM golang:1.21.4-alpine as build
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY . .
RUN go build -o app .

FROM alpine as prod
RUN apk add --no-cache curl grep tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
        echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    rm -rf /var/cache/apk/*

COPY --from=build /build/app /
COPY config.yaml /
CMD ["/app"]
