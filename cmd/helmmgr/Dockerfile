FROM golang:1.13.9-alpine3.11

WORKDIR $GOPATH/src/helm-manager

COPY ./helm3/helm /bin/helm

COPY ./tools/kubectl /bin/kubectl

COPY  ./pkg/ ./pkg/

COPY ./cmd/helmmgr ./cmd/helmmgr

COPY ./vendor/ ./vendor/

COPY ./go.mod ./go.mod

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn

RUN CGO_ENABLED=0 go build -mod=vendor -ldflags '-w -s' -o helmmgr ./cmd/helmmgr

FROM alpine:3.10.3

WORKDIR /helmmgr/app

COPY --from=0 /go/src/helm-manager/helmmgr ./

COPY --from=0 /bin/kubectl /bin/kubectl

COPY --from=0 /bin/helm /bin/helm

WORKDIR /helmmgr

RUN mkdir -p /opt/.kube/

ENTRYPOINT ["./app/helmmgr"]

