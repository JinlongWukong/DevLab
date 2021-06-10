FROM golang as Builder

ENV GO111MODULE=on

WORKDIR /go/src/cloudlab
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o controller .


FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/cloudlab/controller .
COPY --from=builder /go/src/cloudlab/config.ini .
COPY --from=builder /go/src/cloudlab/views/*.html ./views/
COPY --from=builder /go/src/cloudlab/views/image/ ./views/image/
COPY --from=builder /go/src/cloudlab/views/css/ ./views/css/
EXPOSE 8088
ENTRYPOINT ["/app/controller"]
