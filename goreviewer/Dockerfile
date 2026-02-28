FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY goreviewer/go.mod goreviewer/go.sum ./
RUN go mod download

COPY goreviewer/ ./

RUN CGO_ENABLED=0 go build -o goreviewer .

FROM scratch

COPY --from=builder /build/goreviewer /goreviewer

ENTRYPOINT ["/goreviewer"]
