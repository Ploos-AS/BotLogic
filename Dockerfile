FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/botlogic ./cmd/botlogic

FROM alpine:3.22
RUN adduser -D -H -s /sbin/nologin botlogic
COPY --from=build /out/botlogic /usr/local/bin/botlogic
USER botlogic
EXPOSE 8091
ENV BOTLOGIC_LISTEN=0.0.0.0:8091
ENTRYPOINT ["/usr/local/bin/botlogic"]
