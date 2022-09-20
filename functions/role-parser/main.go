package main

import (
	// "context"
	"fmt"

	// "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Users struct {
	Id   string `json:"id"`
	Role string `json:"role"`
}

// func handler(ctx context.Context) (string, error) {
func main() {

	TABLE_NAME := "ofvi"
	INDEX := "list-users"
	KEY_SORT := "SETTINGS"
	OLD_ROLE := "Jefe de área"
	NEW_ROLE := "Aprobador"

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "681334213835_AWSAdministratorAccess",
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
	})
	//Iniciar sesion en aws
	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String(os.Getenv("us-east-1"))},
	// )
	if err != nil {
		panic(fmt.Sprintf("failed to connect session, %v", err))
	}
	svc := dynamodb.New(sess)

	var users []Users

	errQ := svc.QueryPages(&dynamodb.QueryInput{
		TableName:              aws.String(TABLE_NAME),
		IndexName:              aws.String(INDEX),
		KeyConditionExpression: aws.String("sort=:sort"),
		ExpressionAttributeNames: map[string]*string{
			"#role": aws.String("role"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sort": {
				S: aws.String(KEY_SORT),
			},
			":role": {
				S: aws.String(OLD_ROLE),
			},
		},
		FilterExpression: aws.String("#role=:role"),
	}, func(resultQuery *dynamodb.QueryOutput, last bool) bool {
		items := []Users{}

		err := dynamodbattribute.UnmarshalListOfMaps(resultQuery.Items, &items)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		users = append(users, items...)

		return true // keep paging
	})
	if errQ != nil {
		panic(fmt.Sprintf("Got error calling Query: %s", errQ))
	}

	fmt.Print(users)

	for _, req := range users {

		if req.Role == OLD_ROLE {

			input := &dynamodb.UpdateItemInput{
				TableName: aws.String(TABLE_NAME),
				Key: map[string]*dynamodb.AttributeValue{
					"id": {
						S: aws.String(req.Id),
					},
					"sort": {
						S: aws.String(KEY_SORT),
					},
				},
				UpdateExpression: aws.String("set #role = :role"),
				ExpressionAttributeNames: map[string]*string{
					"#role": aws.String("role"),
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
					":role": {
						S: aws.String(NEW_ROLE),
					},
				},
				ReturnValues: aws.String("UPDATED_NEW"),
			}
			_, err := svc.UpdateItem(input)
			if err != nil {
				panic(fmt.Sprintf("failed to Dynamodb Update Items, %v", err))
			}
		}
	}
	// return "", nil
}

// func main() {
// 	lambda.Start(handler)
// }
