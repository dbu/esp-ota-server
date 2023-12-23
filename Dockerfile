FROM golang:alpine AS builder

WORKDIR /build

COPY . .
RUN go build -o espotad ./cmd/espotad

FROM alpine

COPY --from=builder /build/espotad /espotad

ENV EODATADIR /data

VOLUME ["/data"]
EXPOSE 80

ENTRYPOINT ["/espotad"]
