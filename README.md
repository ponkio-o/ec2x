# ec2x

ec2x is cli tool that connect to Amazon EC2 instance easily.

## Install

You can download binary from GitHub Release or build from source. You also need to install `session-manager-plugin` command. If you need more information, please refer to [official document](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

### aqua

This tool supports installation by [aquaproj/aqua](https://github.com/aquaproj/aqua)

```console
$ aqua g -i ponkio-o/ec2x
```

> [!NOTE]  
> If you are using macOS, you can install `session-manager-plugin` command by `aqua g -i aws/session-manager-plugin`.

### GitHub Release

Go to [GitHub Release](https://github.com/ponkio-o/ec2x/releases)

### Build from source

```console
$ go build -o ec2x ./cmd/main.go
$ mv ec2x /usr/local/bin/ec2x
```

## Usage

```console
$ ec2x --help
NAME:
   ec2x - ec2x is connect to EC2 instance using SSM Session Manager

USAGE:
   ec2x [global options] command [command options] [arguments...]

COMMANDS:
   connect  Connect to EC2 instance with Session Manager
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

You can select EC2 instance with fuzzy finder. (using [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder))

```console
$ AWS_PROFILE=<your profile name>
$ ec2x connect

  i-0e70afcef4b54732d - crawler-test (172.19.5.83)                   │  Name            : eks-worker-node
  i-05934f8ec8993dc2c - elasticsearch-prod-01 (172.19.146.230)       │  Architecture    : x86_64
  i-05efb7f9afcbfbca3 - elasticsearch-prod-02 (172.19.155.0)         │  InstanceType    : t3.micro
  i-0d278f748eebdd1f4 - elasticsearch-prod-03 (172.19.180.100)       │  InstanceID      : i-0bb0bade4d8cca310
  i-05838454dd0d2f0f4 - elasticsearch-prod-04 (172.19.181.88)        │  InstanceProfile : eks-node
  i-05223dec50c07cb78 - eks-worker-heavy-01 (172.19.152.122)         │  KeyName         : admin-key
  i-056ebbb9a1f78da01 - sandbox-instance (172.22.202.253)            │  PrivateIP       : 172.22.194.228
  i-035406af724f45017 - es-suggest-v7-prod (172.19.158.223)          │  State           : running
  i-05a762a9bfb78ebd7 - prod-webapp-01 (172.22.202.222)              │
  i-09de276a2e0eaa975 - builder (172.22.207.0)                       │
  i-0c6d6c1dc644c2ef3 - elasticsearch-node-02 (172.19.184.69)        │
  i-0a3d2d7bf7aae6fde - elasticsearch-node-01 (172.19.38.84)         │
  i-0296c5a82bca93012 - sandbox-builder (172.19.34.227)              │
  i-09449ebceb74eaef0 - sandbox-instance (172.19.37.218)             │
> i-0bb0bade4d8cca310 - eks-worker-node (172.22.194.228)             │
  69/69                                                              │
```
