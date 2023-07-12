# Zone trace viewer

## Install

```bash
# Install go and git
sudo apt install git golang -y

# Check go version
# If you see lower than go1.16, you have to install go1.16 or higher version manually.
go version

# Build to /usr/local/bin/
GOPROXY=direct GOBIN=/usr/local/bin go install github.com/pingxiang-chen/bpf-f2fs-zonetrace/viewer@latest
```

## Usage

```bash
f2fs_tracer | viewer
```

If you want to simulate the output of f2fs_tracer, you can use the random input generator:

```bash
python3 random_input.py | viewer
```

## Build from local code

```bash
git clone https://github.com/pingxiang-chen/bpf-f2fs-zonetrace.git
cd bpf-f2fs-zonetrace/viewer
go build
```

 You can run `./viewer`
 