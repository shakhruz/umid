FROM golang:1.14 AS build
WORKDIR /go/src/github.com/umitop/umid/
COPY . .
RUN go build -o umid .

FROM debian:buster-slim
RUN useradd umi
USER umi
WORKDIR /home/umi
COPY --from=build --chown=umi /go/src/github.com/umitop/umid/umid .
CMD ["./umid"]
EXPOSE 8080/tcp
