STACK_NAME ?= blackbox-exporter-lambda
ACCOUNT_ID ?= $(shell aws sts get-caller-identity --output text|cut -f1)
BUCKET     ?= blackbox-exporter-$(ACCOUNT_ID)

LAMBDA_ARN = $(shell aws cloudformation describe-stacks \
						 --output text \
						 --stack-name blackbox-exporter-lambda \
						 --query 'Stacks[0].Outputs[?OutputKey==`LambdaArn`].OutputValue')

ENDPOINT = $(shell aws cloudformation describe-stacks \
					 --output text \
					 --stack-name blackbox-exporter-lambda \
					 --query 'Stacks[0].Outputs[?OutputKey==`Endpoint`].OutputValue')

BUNDLE = handler.zip


require-token:
ifndef AUTH_TOKEN
	$(error AUTH_TOKEN required)
endif

all: $(BUNDLE)

handler:
	go build -o handler

$(BUNDLE): handler
	zip $@ $<

upload: $(BUNDLE)
	aws s3 cp $< s3://$(BUCKET)/$(BUNDLE)

cfn:
	aws cloudformation $(OP) \
  --capabilities CAPABILITY_IAM \
  --stack-name $(STACK_NAME) \
  --parameters ParameterKey=FirstRun,ParameterValue=$(FIRST_RUN) ParameterKey=AuthToken,ParameterValue=$(AUTH_TOKEN) \
  --template-body "$$(cat cfn/site.yml)"

new:
	make cfn OP=create-stack FIRST_RUN=true

update: require-token
	make cfn OP=update-stack FIRST_RUN=false

bump:
	aws lambda update-function-code \
		--function-name "$(LAMBDA_ARN)" \
		--s3-bucket "$(BUCKET)" \
		--s3-key "$(BUNDLE)"

endpoint:
	@echo $(ENDPOINT)
.PHONY: handler cfn new update bump endpoint
