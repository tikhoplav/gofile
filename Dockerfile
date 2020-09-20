FROM golang:alpine as develop
WORKDIR /src
RUN apk add --no-cache git \
	&& go get github.com/cespare/reflex
CMD reflex -r '\.go' -s go run .

FROM golang:alpine as builder
WORKDIR /src
COPY ./src /src
RUN cd /src \
    && apk add --no-cache \
        git \
        gcc \
        musl-dev \
    && go test \
    && go build -o app \
    && cp app /bin/

FROM alpine
COPY --from=builder /bin/app /bin
EXPOSE 80
CMD ["/bin/app"]
