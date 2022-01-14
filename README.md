# Shoptrac

## Building

To build the program, use the default `go build` command. To cross-compile, adjust the Golang compiler settings using the environment variables `GOOS` and `GOARCH`, e.g.:

    GOOS=linux GOARCH=amd64 go build

If you have trouble with libc versions on the target machine, you can instruct Go to build a static binary with Golang-only code by adding the environment variable `CGO_ENABLED=0` to the build command:

    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

