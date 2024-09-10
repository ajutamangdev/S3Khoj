# S3Khoj

[S3Khoj](https://github.com/ajutamangdev/S3Khoj), is a robust tool designed to help detect sensitive files at AWS public S3 buckets. "Khoj", a Nepali word meaning search or explore, perfectly encapsulates the tool's functionality for searching sensitive files within them.

Blog about [S3Khoj](https://csaju.com/posts/hunting-secrets-at-public-s3-buckets-using-s3khoj/).

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
  -d, --download        Download all public files
  -h, --help            help for S3Khoj
  -o, --output string   Output format: text, json, csv, or html (default "text")
  -w, --source string   Custom Wordlist configuration file

```

## Example
```
S3Khoj -b name-of-the-bucket
```

For Custom regex configuration
```
S3Khoj -b name-of-the-bucket -w custom-config.txt
```