# connben

Minimalist TUI network connection benchmarking tool.

# PREVIEW

```
`q` - quit, `h` - hide disconnected.
[ Listening on :8080 | Chunk size 1024 ]

X ->127.0.0.1:52488 224766580_B/s 214.35_MiB/s
X ->127.0.0.1:52496 224036543_B/s 213.66_MiB/s
X ->127.0.0.1:52504 224184021_B/s 213.80_MiB/s
X ->192.168.2.1:48452 118903557_B/s 113.40_MiB/s
->192.168.2.1:53774 86782310_B/s 82.76_MiB/s
->127.0.0.1:52450 79970780_B/s 76.27_MiB/s
```

Disconnected clients are prefixed with `X`.

# INSTALL

```bash
go install github.com/kovmir/connben@latest
```

# USAGE

Simply launch the server:

```bash
connben
# Run `connben -h` to see additional info.
```

Then connect to `connben` server with netcat, which is present on pretty much
any UNIX, or whatever (be sure to specify the right IP and port):

```bash
# Redirect the output to /dev/null to test bandwidth.
nc 127.0.0.1 8080 > /dev/null
# Or redirect to a file to test network along with disk I/O.
nc 127.0.0.1 8080 > /some/directory/file.txt
```

Multiple simultaneous connections are possible, each connection creates two
more threads.
