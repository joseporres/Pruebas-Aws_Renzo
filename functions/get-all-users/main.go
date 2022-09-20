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

type Option struct {
	Title  string `json:"title"`
	Url    string `json:"url"`
	Icon   string `json:"icon"`
	Active bool   `json:"active"`
}

type UserObject struct {
	Id         string   `json:"id"`
	Sort       string   `json:"sort"`
	Name       string   `json:"name"`
	Processes  []Option `json:"processes"`
	Role       string   `json:"role"`
	OfficeRole string   `json:"officeRole"`
	Days       int      `json:"days"`
	Boss       string   `json:"boss,omitempty"`
	BossName   string   `json:"bossName,omitempty"`
}

type Event struct {
	process string
}

func handler(ctx context.Context, event Event) ([]UserObject, error) {

	TABLE_NAME := os.Getenv("TABLA_NAME")
	INDEX := "list-users"
	SORT := "SETTINGS"
	var usersData []UserObject
	var users []UserObject

	var process string
	switch event.process {
	case "SW":
		process = "Smart Working"
	case "OF":
		process = "Oficios"
	default:
		process = ""
	}

	//Iniciar sesion en aws
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION"))},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect session, %v", err))
	}
	svc := dynamodb.New(sess)

	errQ := svc.QueryPages(&dynamodb.QueryInput{
		TableName:              aws.String(TABLE_NAME),
		IndexName:              aws.String(INDEX),
		KeyConditionExpression: aws.String("sort = :sort"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sort": {
				S: aws.String(SORT),
			},
		},
		ProjectionExpression: aws.String("id"),
	}, func(resultQuery *dynamodb.QueryOutput, last bool) bool {
		items := []UserObject{}
		err := dynamodbattribute.UnmarshalListOfMaps(resultQuery.Items, &items)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		usersData = append(users, items...)

		return true // keep paging
	})

	if errQ != nil {
		panic(fmt.Sprintf("Got error calling Query: %s", errQ))
	}

	if process == "" {
		users = usersData
	} else {
		for _, usr := range usersData {
			for _, usrPrcss := range usr.Processes {
				if usrPrcss.Title == process && usrPrcss.Active {
					users = append(users, usr)
					break
				}
			}
		}
	}

	return users, nil
}

func main() {
	lambda.Start(handler)
}
