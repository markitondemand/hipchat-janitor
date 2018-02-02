FROM alpine:latest
COPY janitor /usr/bin/janitor
EXPOSE 3000
EXPOSE 3001
ENTRYPOINT ["/usr/bin/janitor"]
