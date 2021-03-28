# afl-transmit

Transfer AFL files over a mesh to fuzz across multiple servers 

## Usage

You need to specify your AFL output directory with `--fuzzer-directory`, and your peers with `--peers`.
Some other options exist to let you fine-tune your *afl-transmit* experience, have a look at them via `--help`.

On default, *afl-transmit* opens port 1337/TCP to wait for incoming connections. If you are not on a private net, make sure to protect this port with a firewall, or anyone on the internet may send you files (although this might become interesting).

### Quickstart

- On your host 10.0.0.1: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.2,10.0.0.3`
- On your host 10.0.0.2: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1,10.0.0.3`
- On your host 10.0.0.3: `./afl-transmit --fuzzer-directory /ram/output --peers 10.0.0.1,10.0.0.2`

