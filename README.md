# afl-transmit

Transfer AFL files over a mesh to fuzz across multiple servers 

## Features

- Using DEFLATE compression format (see [RFC 1951](https://www.ietf.org/rfc/rfc1951.html))
- Automatically syncs the main fuzzer to secondary nodes, and all secondary fuzzers back to the main node
- Encrypts traffic between nodes using AES-256, dropping plaintext packets
- Usable on UNIXoid (Linux, OSX) systems and Windows
- Reduces the amount of transmitted test cases to a bare minimum

## Usage

You need to specify your AFL output directory with `--fuzzer-directory`, and your peers with `--peers`.
Some other options exist to let you fine-tune your *afl-transmit* experience, have a look at them via `--help`.

On default, *afl-transmit* opens port 1337/TCP to wait for incoming connections. If you are not on a private net, make sure to protect this port with a firewall, or anyone on the internet may send you files (although this might become interesting).
As a countermeasure, use the `--restrict-to-peers` flags to only allow connections from your known peers.

### Quickstart

- On your host 10.0.0.1: `./afl-transmit --fuzzer-directory /ram/output --main --peers 10.0.0.2,10.0.0.3`
- On your host 10.0.0.2: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1`
- On your host 10.0.0.3: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1`

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

### Traffic reduction

On default, *afl-transmit* avoids sending files with the same file present in different fuzzer directories.
This will greatly reduce the traffic between your nodes (I measured 621 kB to 1.3 kB, for example).
Please note that there might be some edge cases when you don't want that behaviour, e.g.
- you want to preserve the queue of each fuzzer
- you expect your fuzzers to give the same (file) name to different test cases, in which case *afl-transmit* would mistakenly assume that the file has the same *contents* and not only the same *name*
- you don't care for traffic

To avoid reducing the transmitted files by comparing filenames, add `--no-duplicates=false` as argument.

Also on default, *afl-transmit* tries to check if the queue of the observed fuzzers contain test cases which originated from another fuzzer instance.
In that case, the file name contains the keyword "sync" in it, and looks e.g. like this: `id:001815,time:0,orig:id:001805,sync:main,src:001794`
If it was copied from another fuzzer, it means that the file is already present in the fuzzer cluster, and can safely be skipped on those fuzzer instances which copied it.
Please note that this will produce false positives if the filename of your testcases contain `,sync:` for whatever reason.

To avoid reducing the transmitted files by filtering synced files out, add `--avoid-synced=false` as argument.

If you still have trouble paying the invoice for your ISP due to heavy traffic usage, try increasing the `--rescan` value, so files are transmitted less often.
