FROM golang as Builder

ENV GO111MODULE=on

WORKDIR /go/src/devlab
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o controller .


FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/devlab/controller .
COPY --from=builder /go/src/devlab/config.ini .
COPY --from=builder /go/src/devlab/views/*.html ./views/
COPY --from=builder /go/src/devlab/views/image/ ./views/image/
COPY --from=builder /go/src/devlab/views/css/ ./views/css/
COPY --from=builder /go/src/devlab/views/scripts/ ./views/scripts/
RUN mkdir .db/
EXPOSE 8088
ENTRYPOINT ["/app/controller"]
