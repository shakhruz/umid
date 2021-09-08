FROM golang:1.17 AS build
WORKDIR /go/src/github.com/umitop/umid/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -trimpath -mod vendor -ldflags '-s -w -extldflags "-static"' -tags netgo -o umid .
RUN echo 'umi:x:1000:1000::/home/umi:/bin/sh' > /etc/passwd_

FROM scratch
COPY --from=build /etc/passwd_ /etc/passwd
COPY --from=build --chown=1000:1000 /go/src/github.com/umitop/umid/umid /umid
USER umi
CMD ["/umid"]
EXPOSE 8080/tcp
