FROM golang:1.15-buster AS build

ARG VERSION=undefinied

WORKDIR /src/
COPY cmd/ /src/cmd/
COPY internal/ /src/internal/
COPY go.mod /src/
COPY go.sum /src/
COPY Makefile /src/
# copy git for the tag
#COPY .git /src/.git/
WORKDIR /src/
#ENV GOROOT=/
ENV GO111MODULE=on
RUN make build-linux64

FROM debian:10.9
COPY --from=build /src/bin/repow_linux-amd64 /bin/repow
ENTRYPOINT ["/bin/repow"]
CMD ["serve"]
