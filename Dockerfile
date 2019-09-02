FROM golang:1.12-alpine AS builder
RUN apk add --no-cache git

WORKDIR /csgo-roster
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go install ./cmd/app

FROM alpine
RUN apk add --no-cache ca-certificates git
COPY --from=builder /go/bin /bin
USER nobody:nobody
ENTRYPOINT [ "/bin/csgo-roster" ]