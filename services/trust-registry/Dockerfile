ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
ENV GODEBUG=http2client=0
WORKDIR /src

COPY . .
RUN for attempt in 1 2 3; do \
      timeout 180s sh -c 'CGO_ENABLED=0 go build -trimpath -o /out/trust-registry ./services/trust-registry/cmd/trust-registry' && exit 0; \
      sleep 5; \
    done; \
    exit 1

FROM alpine:3.20
RUN addgroup -S hdip && adduser -S -G hdip hdip
USER hdip

COPY --from=build /out/trust-registry /usr/local/bin/trust-registry
EXPOSE 8080
ENTRYPOINT ["trust-registry"]
