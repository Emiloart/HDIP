ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS build
ENV GODEBUG=http2client=0
WORKDIR /src

COPY . .
RUN for attempt in 1 2 3; do \
      timeout 180s sh -c 'CGO_ENABLED=0 go build -trimpath -o /out/phase1sql ./services/internal/phase1sql/cmd/phase1sql' && exit 0; \
      sleep 5; \
    done; \
    exit 1

FROM alpine:3.20
RUN addgroup -S hdip && adduser -S -G hdip hdip
USER hdip

COPY --from=build /out/phase1sql /usr/local/bin/phase1sql
ENTRYPOINT ["phase1sql"]
