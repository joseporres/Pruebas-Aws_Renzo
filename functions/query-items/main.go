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

type UsersId struct {
	UserId string `json:"sort"`
}

func handler(ctx context.Context) (string, error) {

	TABLE_NAME := os.Getenv("TABLA_NAME")
	var usersId []UsersId

	//Iniciar sesion en aws
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("us-east-1"))},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect session, %v", err))
	}
	svc := dynamodb.New(sess)

	errQ := svc.QueryPages(&dynamodb.QueryInput{
		TableName:              aws.String(TABLE_NAME),
		KeyConditionExpression: aws.String("id= :id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String("SETTINGS"),
			},
		},
		ProjectionExpression: aws.String("sort"),
	}, func(resultQuery *dynamodb.QueryOutput, last bool) bool {
		items := []UsersId{}
		err := dynamodbattribute.UnmarshalListOfMaps(resultQuery.Items, &items)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		usersId = append(usersId, items...)

		return true // keep paging
	})
	if errQ != nil {
		panic(fmt.Sprintf("Got error calling Query: %s", errQ))
	}
	fmt.Println("el query retorna: ", usersId)

	for _, usr := range usersId {

		input := &dynamodb.UpdateItemInput{
			TableName: aws.String(TABLE_NAME),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String("SETTINGS"),
				},
				"sort": {
					S: aws.String(usr.UserId),
				},
			},
			UpdateExpression: aws.String("set newAttr = :newAttr"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":newAttr": {
					S: aws.String("nuevoAttr"),
				},
			},
			ReturnValues: aws.String("UPDATED_NEW"),
		}
		_, err := svc.UpdateItem(input)
		if err != nil {
			panic(fmt.Sprintf("failed to Dynamodb Update Items, %v", err))
		}
	}

	return "", nil
}

func main() {
	lambda.Start(handler)
}
