# builder image
FROM golang:1.13-alpine3.11 as builder

WORKDIR /encrypted-secrets
COPY . /encrypted-secrets
RUN go build -o /bin/encrypted-secrets ./cmd/manager/main.go

# final image
FROM alpine:3.11

RUN apk --no-cache add ca-certificates dumb-init
COPY --from=builder /bin/encrypted-secrets /bin/encrypted-secrets

USER 65534
ENTRYPOINT ["dumb-init", "--", "/bin/encrypted-secrets"]