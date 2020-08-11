FROM golang:1.14 AS build
WORKDIR /go/src/github.com/umitop/umid/
COPY . .
RUN go build -ldflags "-extldflags -static" -tags netgo -o umid .
RUN echo 'umi:x:1000:1000::/home/umi:/bin/sh' > /etc/passwd2

FROM alpine AS alp

FROM scratch
#COPY --from=alp /bin/ls /bin/ls
#COPY --from=alp /bin/sh /bin/sh
#COPY --from=alp /usr/bin/ldd /usr/bin/ldd
#COPY --from=alp /usr/bin/whoami /usr/bin/whoami
#COPY --from=alp /lib/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY --from=build /etc/passwd2 /etc/passwd
COPY --from=build --chown=1000:1000 /go/src/github.com/umitop/umid/umid /umid
USER umi
CMD ["/umid"]
EXPOSE 8080/tcp
