package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/reminder-serverless-app/reminder-serverless-app-todos/handlers"
)

func main() {
	lambda.Start(handlers.HandleRequest)
}
