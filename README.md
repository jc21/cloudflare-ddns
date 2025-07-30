# Cloudflare DDNS

- Sick of using Dynamic DNS services that are costly or annoying?
- Already use Cloudflare for your domain DNS?
- Do you have a dynamically assigned IP address?

This is the command for you.

```bash
Update Cloudlfare DNS record with your current IP address
Usage: cloudflare-ddns [--setup] [--configfile CONFIGFILE]

Options:
  --setup, -s            Setup wizard
  --configfile CONFIGFILE, -c CONFIGFILE
                         Config File to use (default: ~/.config/cloudflare-ddns.json)
  --help, -h             display this help and exit
  --version              display version and exit
```

## Install

#### RHEL based distributions

RPM hosted on [yum.jc21.com](https://yum.jc21.com)

#### Go install

```bash
go install github.com/jc21/cloudflare-ddns@latest
```

#### Building

```bash
git clone https://github.com/jc21/cloudflare-ddns && cd cloudflare-ddns
go build -o bin/cloudflare-ddns cmd/cloudflare-ddns/main.go
./bin/cloudflare-ddns -h
```
