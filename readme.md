# SNI Bypass

A python script which generate self signed SSL certs, and use caddy as MITM to bypass SNI filter.

## requirements:

1. `python 3.7` and above
2. `pyOpenSSL`
3. [`caddyserver v2`](https://caddyserver.com/download)

## Usage:

simply run `python main.py`, and it will read the config from `config.json`, then do the jobs.

## Attention

Some upstreams are also reverse proxy, if you don't know what they are doing, please just disbale these upstreams.
