package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ProcessRequest struct {
	ID         string `json:"id"`
	Sort       string `json:"sort"`
	QuizAnswer string `json:"quizAnswer"`
	QuizTime   string `json:"quizTime"`
	State      int    `json:"state"`
}

type QuizAnswer struct {
	Question string `json:"question"`
	Answer   bool   `json:"answer"`
}

func handler(ctx context.Context) (string, error) {

	TABLE_NAME := "ofvi-Officio"
	INDEX := "id-state"
	KEY_ID := "PROCESS_REQUEST"
	KEY_STATE := "8"

	QUESTIONS := make([]string, 5)
	QUESTIONS[0] = "Pregunta 1"
	QUESTIONS[1] = "Pregunta 2"
	QUESTIONS[2] = "Pregunta 3"
	QUESTIONS[3] = "Pregunta 4 old"
	QUESTIONS[4] = "Pregunta 4 new"

	var requests []ProcessRequest

	layout := "2006-01-02T15:04:05.000Z"
	date, err := time.Parse(layout, "2022-06-24T00:00:00.000Z")
	if err != nil {
		panic(fmt.Sprintf("failed to parse to Time, %v", err))
	}

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
		IndexName:              aws.String(INDEX),
		KeyConditionExpression: aws.String("id = :id AND #state=:state"),
		ExpressionAttributeNames: map[string]*string{
			"#state": aws.String("state"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String(KEY_ID),
			},
			":state": {
				N: aws.String(KEY_STATE),
			},
		},
	}, func(resultQuery *dynamodb.QueryOutput, last bool) bool {
		items := []ProcessRequest{}

		err := dynamodbattribute.UnmarshalListOfMaps(resultQuery.Items, &items)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal Dynamodb Scan Items, %v", err))
		}

		requests = append(requests, items...)

		return true // keep paging
	})
	if errQ != nil {
		panic(fmt.Sprintf("Got error calling Query: %s", errQ))
	}

	for _, req := range requests {

		var newAnswers []QuizAnswer
		var oldAnswers []bool

		errUnmarshall := json.Unmarshal([]byte(req.QuizAnswer), &oldAnswers)
		if errUnmarshall != nil {
			//panic(fmt.Sprintf("Got error when try unmarshall string: %s", errUnmarshall))
			continue
		}
		for i := 0; i < 4; i++ {
			if i == 3 {
				dateQuiz, err := time.Parse(layout, req.QuizTime)
				if err != nil {
					panic(fmt.Sprintf("failed to parse to Time, %v", err))
				}
				if dateQuiz.Before(date) {
					newAnswers = append(newAnswers, QuizAnswer{QUESTIONS[i], oldAnswers[i]})
				} else {
					newAnswers = append(newAnswers, QuizAnswer{QUESTIONS[i+1], oldAnswers[i]})
				}
			} else {
				newAnswers = append(newAnswers, QuizAnswer{QUESTIONS[i], oldAnswers[i]})
			}
		}

		newAnswersStr, errMarshall := json.Marshal(newAnswers)
		if errMarshall != nil {
			panic(fmt.Sprintf("Got error when try marshall json: %s", errMarshall))
		}

		input := &dynamodb.UpdateItemInput{
			TableName: aws.String(TABLE_NAME),
			Key: map[string]*dynamodb.AttributeValue{
				"id": {
					S: aws.String(KEY_ID),
				},
				"sort": {
					S: aws.String(req.Sort),
				},
			},
			UpdateExpression: aws.String("set quizAnswer = :quizAnswer"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":quizAnswer": {
					S: aws.String(string(newAnswersStr)),
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
