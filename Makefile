# e.g. profile_abc
AWS_PROFILE:=aws_profile_name
# e.g. us-east-1
AWS_REGION:region_name
# e.g. staging or prod	
ENV_NAME:=env_name

install_deps:
	go get -d ./...
clean: 
	rm -rf ./populate-to-dynamodb/populate-to-dynamodb
build:
	GOOS=linux GOARCH=amd64 go build -o populate-to-dynamodb/populate-to-dynamodb ./populate-to-dynamodb
run:
	sam local start-lambda
create_code_bucket:
	aws s3 mb s3://populate-data-code --region $(AWS_REGION) --profile $(AWS_PROFILE)		
create_csv_bucket:
	aws s3 mb s3://populate-data-csv --region $(AWS_REGION) --profile $(AWS_PROFILE)		
package:
	sam package --output-template-file packaged.yaml --s3-bucket populate-data-code --profile $(AWS_PROFILE) --region $(AWS_REGION)
deploy:
	sam deploy --template-file packaged.yaml --stack-name sam-test-app-nirdosh-$(ENV_NAME) --capabilities CAPABILITY_IAM --profile $(AWS_PROFILE) --region $(AWS_REGION)