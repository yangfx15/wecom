# 临时构建容器，建议源为一个干净或带了通用依赖的基础镜像
FROM registry01.wezhuiyi.com/library/golang:1.17.4-centos7.6 as builder

# TODO: 根据实际情况，声明编译所需要的环境变量
ENV GOPROXY http://pkg.in.wezhuiyi.com/repository/golang/,direct
ENV GOPRIVATE code.in.wezhuiyi.com
ENV GOSUMDB off
ENV GO111MODULE on

WORKDIR /build

COPY . .

RUN go mod download && go get ./... && go mod tidy && CGO_ENABLED=0 go build -o wecom main.go

FROM registry01.wezhuiyi.com/library/alpine:3.10
ENV WECOM=/app
WORKDIR /app

COPY --from=builder /build/wecom /app/wecom
#似乎不需要此行也会创建logs目录
#RUN mkdir ${WECOM}/logs

ENTRYPOINT ["/app/wecom"]
