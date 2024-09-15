# S3Khoj

<p align="center">
<img src="https://img.shields.io/github/go-mod/go-version/ajutamangdev/s3khoj">
<a href="https://github.com/ajutamangdev/s3khoj/releases"><img src="https://img.shields.io/github/downloads/ajutamangdev/s3khoj/total">
<a href="https://github.com/ajutamangdev/s3khoj"><img src="https://img.shields.io/github/release/ajutamangdev/s3khoj">
<a href="https://github.com/ajutamangdev/s3khoj/issues"><img src="https://img.shields.io/github/issues-raw/ajutamangdev/s3khoj">
<a href="https://github.com/ajutamangdev/s3khoj/discussions"><img src="https://img.shields.io/github/discussions/ajutamangdev/s3khoj">
      
[S3Khoj](https://github.com/ajutamangdev/S3Khoj), is a robust tool designed for pentesters to extract juicy information from the public accessible S3 buckets. "Khoj", a Nepali word meaning search or explore, perfectly encapsulates the tool's functionality for searching sensitive files within them.

Blog about [S3Khoj](https://csaju.com/posts/hunting-secrets-at-public-s3-buckets-using-s3khoj/).

## Installation

Manual
```
git clone https://github.com/ajutamangdev/S3Khoj
cd S3Khoj
make build
./S3Khoj -h
```
> Ensure you have installed go in your machine for the build process.

Build S3khoj uusing Docker locally
```
docker build -t S3Khoj .
```

Pull S3Khoj docker image using DockerHub
```
docker pull ajutamangdev/s3khoj 
```

You can also download the binary from https://github.com/ajutamangdev/S3Khoj/releases and installed on your machine.

## Usage

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

If you are running from Docker, you have to mount the volumes.
```
docker run -v $(pwd):/app -w /app s3 -b test1011hify -o html
```

For Custom regex configuration
```
S3Khoj -b name-of-the-bucket -w custom-config.txt
```

### License

S3Khoj is distributed under [MIT License](https://github.com/ajutamangdev/S3Khoj/blob/main/LICENSE)
