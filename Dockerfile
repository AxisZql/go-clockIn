FROM golang:alpine as builder
WORKDIR /home
# 使用sqlite3 driver需要用到cgo，所以需要安装gcc
RUN echo -e http://mirrors.aliyun.com/alpine/v3.10/main/ > /etc/apk/repositories \
    && apk update \
    && apk add --no-cache gcc g++ libffi-dev musl-dev openssl-dev make linux-headers libc-dev libc6-compat binutils
COPY . .
RUN go build --mod=vendor -o gzhuClockIn ./internal/cmd/main.go

FROM alpine as runner
ENV WORKDIR=/home
WORKDIR $WORKDIR
RUN apk add tzdata --no-cache \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata
COPY --from=builder $WORKDIR/config/ $WORKDIR/config
COPY --from=builder $WORKDIR/gzhuClockIn $WORKDIR
COPY --from=builder $WORKDIR/db.sqlite3 $WORKDIR/db.sqlite3
CMD ["./gzhuClockIn"]