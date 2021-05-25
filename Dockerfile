FROM golang as Builder

ENV GO111MODULE=on

WORKDIR /go/src/cloudlab
COPY . .
RUN go build -o controller .


FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/cloudlab/controller .
COPY --from=builder /go/src/cloudlab/config.ini .
EXPOSE 8088
ENTRYPOINT /app/controller 2>&1