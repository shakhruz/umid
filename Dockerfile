FROM golang:1.17 AS build
WORKDIR /go/src/gitlab.com/umitop/umid/
COPY ./cmd ./cmd
COPY ./pkg ./pkg
COPY ./go.mod ./go.mod
# go build [-o output] [build flags] [packages]
# -trimpath
#     remove all file system paths from the resulting executable.
# -ldflags '[pattern=]arg list'
#     arguments to pass on each go tool link invocation.
# -tags tag,list
#     a comma-separated list of build tags to consider satisfied during the
#	  build. For more information about build tags, see the description of
#	  build constraints in the documentation for the go/build package.
RUN CGO_ENABLED=0 go build -o umid -trimpath -ldflags "-s -w -extldflags '-static'" -tags netgo ./cmd/umid
RUN echo 'umi:x:1000:1000::/home/umi:/bin/sh' > /etc/passwd_

FROM scratch
COPY --from=build /etc/passwd_ /etc/passwd
COPY --from=build --chown=1000:1000 /go/src/gitlab.com/umitop/umid/umid /umid
USER umi
ENV UMI_BIND=:8080
WORKDIR /home/umi
CMD ["/umid"]
EXPOSE 8080/tcp
