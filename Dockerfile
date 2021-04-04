FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

ENV USER=appuser
ENV UID=10001 

RUN adduser \    
    --disabled-password \    
    --gecos "" \    
    --home "/nonexistent" \    
    --shell "/sbin/nologin" \    
    --no-create-home \    
    --uid "${UID}" \    
    "${USER}"
WORKDIR $GOPATH/src/open-graph-service
COPY . .

RUN go get -d -v

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /go/bin/open-graph-service

FROM scratch

EXPOSE 8080

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/open-graph-service /go/bin/open-graph-service

USER appuser:appuser

ENTRYPOINT ["/go/bin/open-graph-service"]