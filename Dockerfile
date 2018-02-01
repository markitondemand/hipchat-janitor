FROM alpine:latest
COPY janitor /usr/bin/janitor
ENTRYPOINT ["/usr/bin/janitor"]
