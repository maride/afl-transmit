# afl-transmit

Transfer AFL files over a mesh to fuzz across multiple servers 

## Features

- Automatically syncs the fuzzers over all nodes
- No obscure dependencies, no painful setup process - just a single, self-contained binary
- Using DEFLATE compression format (see [RFC 1951](https://www.ietf.org/rfc/rfc1951.html))
- Encrypts traffic between nodes using AES-256, dropping plaintext packets
- Usable on UNIX-like systems (Linux, OSX) and Windows

## Usage

You need to specify your AFL output directory with `--fuzzer-directory`, and your peers with `--peers`.
Some other options exist to let you fine-tune your *afl-transmit* experience, have a look at them via `--help`.

On default, *afl-transmit* opens port 1337/TCP to wait for incoming connections. If you are not on a private net, make sure to protect this port with a firewall, or anyone on the internet may send you files (although this might become interesting).
As a countermeasure, use the `--restrict-to-peers` flags to only allow connections from your known peers.

### Quickstart

Let's assume you have three servers running with some instances of AFL, all in secondary (`-S`) mode, except the main fuzzer running on the box 10.0.0.1.
To sync test cases across those servers, you'd need to run
- on 10.0.0.1: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.2,10.0.0.3`
- on 10.0.0.2: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1,10.0.0.3`
- on 10.0.0.3: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1,10.0.0.2`

Because *afl-transmit* stays in the foreground, you should probably run it in a `tmux` window or something comparable.

### Crypto

If you want to encrypt your traffic between the nodes - which is advised, as it increases security and there is nearly no argument against it - you can do so by specifying a random key with `--key`.
To keep *afl-transmit* simple, the symmetric encryption algorithm AES256-GCM was chosen over an asymmetric variant. This means you need to specify the same key on all nodes.

Key generation is fairly simple, you just need to get 32 random bytes from somewhere (buy them, or use `/dev/urandom`), and wrap them with base64.
For example like this:

```
dd if=/dev/urandom bs=32 count=1 2>/dev/null | base64 | tee transmit.key
./afl-transmit --key $(cat transmit.key) --fuzzer-directory ...
```

As already said, the same key must be used on all nodes.
