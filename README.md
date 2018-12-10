# Lambda
This project allows you to run the [Prometheus Blackbox
Exporter](https://github.com/prometheus/blackbox_exporter) as [AWS
Lambda](https://aws.amazon.com/lambda/) behind the [AWS API
Gateway](https://aws.amazon.com/api-gateway/).

It uses a [Cloudformation](https://aws.amazon.com/cloudformation/) template to
create a S3 bucket (named `blackbox-exporter-[AWS AccountId]`).

## Deploy Stack
Since the code needs to exist when we create the lambda, on first run we disable
it's creation by setting FirstRun=true:

### Deploy stack w/o Lambda
```
aws cloudformation create-stack \
  --capabilities CAPABILITY_IAM \
  --stack-name test-lambda \
  --parameters ParameterKey=FirstRun,ParameterValue=true \
  --template-body "$(cat cfn/site.yml)"
```

### Upload Lambda
```
make upload
```


### Update stack to enable Lambda
```
aws cloudformation update-stack \
  --capabilities CAPABILITY_IAM \
  --stack-name test-lambda \
  --template-body "$(cat cfn/site.yml)"
```
