FROM golang:1.10.2-alpine as builder

RUN mkdir -p /go/src/github.com/adbourne/website-seacitysoftware/

WORKDIR /go/src/github.com/adbourne/website-seacitysoftware/

COPY . /go/src/github.com/adbourne/website-seacitysoftware/

RUN apk add --update make git &&\
    ls -l &&\
    export GOPATH="/go" &&\
    make buildTools &&\
    make dependenciesBackend &&\
    make buildBackendLinux BACKEND_DIR=website-sea-city-software

##
# Build the final container
##
FROM alpine:3.8
MAINTAINER Aaron Bourne <contact@aaronbourne.co.uk>

ENV FRONTEND_DIR="/views"

WORKDIR /

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Create an app user to run the application as
RUN addgroup -S app && adduser -S -g app app

# Add the static files
COPY --from=builder /go/src/github.com/adbourne/website-seacitysoftware/views /views

COPY --from=builder /go/src/github.com/adbourne/website-seacitysoftware/public /public

# Add the built application
COPY --from=builder /go/src/github.com/adbourne/website-seacitysoftware/target/website-sea-city-software /website-sea-city-software

RUN chown app:app /website-sea-city-software &&\
    chown -R app:app /views &&\
    chown -R app:app /public &&\
    chmod +x /website-sea-city-software

# Run as the app user
USER app

# Run the app
ENTRYPOINT ["/website-sea-city-software"]