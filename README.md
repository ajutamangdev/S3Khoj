# S3Khoj

S3Khoj, S3 inspector tool that help pentesters to extract juicy information from the public accessible S3 buckets.

# Installation

```
git clone https://github.com/ajutamangdev/S3Khoj
cd S3Khoj
make install
```
> Ensure you have installed go in your machine.

# Usage

You can check with the help flag by executing the given command.
```
> S3Khoj -h
S3Khoj is a inspector tool that help pentesters to extract juicy information from the public accessible S3 buckets.

Usage:
  S3Khoj [flags]

Flags:
  -b, --bucket string   Name of the s3 bucket to check
  -h, --help            help for S3Khoj
  -s, --source string   External directory list file

```

## Example
```
S3Khoj -b name-of-the-bucket
```

For Custom regex configuration
```
S3Khoj -b name-of-the-bucket -s custom-config.txt
```