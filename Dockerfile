ARG GOVERSION=alpine
ARG TARGETOS
ARG TARGETARCH

FROM --platform=$BUILDPLATFORM golang:$GOVERSION AS build
WORKDIR /src
COPY . /src
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/mdbin

FROM alpine
COPY --from=build /bin/mdbin /bin/mdbin
EXPOSE 23342
VOLUME /html
CMD ["mdbin", "serve", "--port", "23342", "--htmldir", "/html"]
