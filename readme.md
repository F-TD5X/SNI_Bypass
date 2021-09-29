# SNI Bypass

[Python + Caddy](https://github.com/F-TD5X/SNI_Bypass/tree/main)

[Go](https://github.com/F-TD5X/SNI_Bypass/tree/go)

## Why Go version

Caddy is good, but too heavy for just a reverse proxy.

## Usage

1. run the `.exe` file.
2. trust the `CA.crt` generated into the folder.
3. leave it running, it's ok to go.

## requirements

A system can use TLS and enough resources for this program.

## Build

1. Clone this branch.
2. `go build`


# TODO
- [ ] better certificate management
- [ ] upstream choose, now only round_robin
- [ ] better config file and parse progress
- [ ] Optimize the main code