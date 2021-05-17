# ncdu-s3

ncdu-s3 records file sizes of S3 buckets to be used with ncdu.

## Installation and Usage

```console
$ go get github.com/akrennmair/ncdu-s3
$ ncdu-s3 s3://bucket-name/path-prefix/ output.json
$ ncdu -f output.json
```

## Configuration

By default, `ncdu-s3` will use your default AWS credential configuration in
`~/.aws`. You can override this by setting the environment variables
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` with the appropriate
credentials. If you have multiple profiles configured in your `~/.aws`,
you can also use an alternative profile by setting the `AWS_PROFILE`
environment variable. Most likely, you will also have to set the
`AWS_REGION` environment variable.

## License

See the file `LICENSE.md` for the license.

## Author

Andreas Krennmair <ak@synflood.at>