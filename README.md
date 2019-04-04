[![Stability](https://img.shields.io/badge/Stability-Under%20Active%20Development-Red.svg)](https://github.com/pbs/redyl)

# Warning: experimental

This is an experimental library, and is currently unsupported

`redyl` handles multi-factor authentication and automatic IAM key rotation for human use of the AWS API.


# Usage

```
$ redyl
```

You should receive a prompt for an MFA code. If it's a good code, `redyl` will write rotate your access keys and write updated values to ~/.aws/credentials.

# Requirements

You'll need your credentials files set up in a particular way for this to work:

```
$ cat ~/.aws/config

[default]
region = us-east-1
# you can find this under Assigned MFA device at https://console.aws.amazon.com/iam/home?#/users/youruser?section=security_credentials 
mfa_serial = arn:aws:iam::account_number:mfa/username 

```

```
$ cat ~/.aws/credentials

# use default_original here, not default - redyl will write temporary credentials to default
[default_original]
aws_access_key_id     = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
```

# Installation

1) Find the latest release on https://github.com/pbs/redyl/releases and download the binary for your system (look for `darwin` if you're using a mac, `linux` or `windows` otherwise).

2) Move the binary somewhere in your PATH

`mv ~/Downloads/redyl-0.1.0-darwin-amd64 /usr/local/bin/redyl`

3) Set the executable bit

`chmod +x /usr/local/bin/redyl`

4) `redyl` should now work, try `redyl version`
