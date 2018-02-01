# HipChat Janitor

Automatically archive unused private HipChat rooms.

## Usage

The easiest way to use this is the Docker container:

```bash
$ docker run -d hipchat-janitor -token=<HipChat API Token> -url=<HipChat server URL>
```

This will cleanup any rooms that haven't been touched in 30 days every 24 hours.

## Command-Line Flags

```
  -insecure
        Skip certificate verification for HTTPS requests.
  -interval int
        How often cleanups are attempted, in hours. (default 24)
  -max int
        The maximum amount of time to keep a room, in days. (default 30)
  -token string
        The HipChat API token.
  -url string
        The HipChat server URL.
```
