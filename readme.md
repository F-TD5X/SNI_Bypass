# SNI Bypass

[Python + Caddy](https://github.com/F-TD5X/SNI_Bypass/tree/main)

[Go](https://github.com/F-TD5X/SNI_Bypass/tree/go)

## Why Go version

Caddy is good, but too heavy for just a reverse proxy.

## Usage

1. run the `.exe` file.
2. edit the `config.yml` to fit your needs.
3. trust the `CA.crt` generated into the folder.
4. leave it running, it's ok to go.

## requirements

A system can use TLS and enough resources for this program.

## Build

1. Clone and switch to this branch.
2. `go build -ldflags '-s -w'`


# Known issue

1. `git` and some cli tools has certificate problem.

Because they are build with `ca-certificates` and sometimes not trust your self-signed certificate. You have to manually ignore this errors. Like `git -c http.sslVerify=false` or `GIT_SSL_NO_VERIFY=true git `

2. steamcommunity.com is too slow.

You known, steamcomunity's videos are hosted on youtube.  You will have problem there if you can't access youtube. This program can't help you bypass TCP RST.

# TODO
- [ ] Auto register as a service.
- [ ] better certificate management
- [ ] upstream choose, now only round_robin
- [ ] better config file and parse progress
- [ ] Optimize the main code