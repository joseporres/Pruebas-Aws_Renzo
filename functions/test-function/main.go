package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Item struct {
	Sort       string `json:"sort"`
	AreaEntity string `json:"areaEntity"`
	AreaName   string `json:"areaName"`
	Name       string `json:"name"`
}

type Event struct {
	Id string `json:"id"`
}

func handler(ctx context.Context, event Event) (string, error) {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("us-east-1"))},
	)
	if err != nil {
		return "", err
	}
	svc := dynamodb.New(sess)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("TablaPrueba"),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String("AREA"),
			},
			"sort": {
				S: aws.String(event.Id),
			},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("Got error calling GetItem, %s", err))
	}

	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		panic(fmt.Sprintf("Got error calling GetItem, %s", err))
	}

	fmt.Println(item)

	return "Success", nil
}

func main() {
	lambda.Start(handler)
}
