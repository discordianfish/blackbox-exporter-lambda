# Lambda
This project allows you to run the [Prometheus Blackbox
Exporter](https://github.com/prometheus/blackbox_exporter) as [AWS
Lambda](https://aws.amazon.com/lambda/) behind the [AWS API
Gateway](https://aws.amazon.com/api-gateway/).

It uses a [Cloudformation](https://aws.amazon.com/cloudformation/) template to
create a S3 bucket (named `blackbox-exporter-[AWS AccountId]`).

It requires a `Authorization: Bearer xx` token that is set by the `AUTH_TOKEN`
environment variable.

## Deploy Stack
Since the code needs to exist when we create the lambda, on first run we disable
it's creation by setting FirstRun=true:

Deploy stack w/o Lambda:
```
make new
```

Upload Lambda:
```
make upload
```

Update stack to enable Lambda:
```
AUTH_TOKEN=xx make update
```
