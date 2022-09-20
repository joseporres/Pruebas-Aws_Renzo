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

	TABLE_NAME := os.Getenv("TABLA_NAME")
	INDEX := "id-state"
	KEY_ID := "PROCESS_REQUEST"
	KEY_STATE := "8"

	QUESTIONS := make([]string, 10)
	QUESTIONS[0] = "Sensación de alza térmica o fiebre."
	QUESTIONS[1] = "Dolor de garganta, tos, estornudos o dificultad para respirar."
	QUESTIONS[2] = "Expectoración o flema amarilla o verdosa."
	QUESTIONS[3] = "Pérdida del gusto y/o olfato."
	QUESTIONS[4] = "Contacto con persona(s) con un caso confirmado de COVID-19."
	QUESTIONS[5] = "Está tomando alguna medicación."
	QUESTIONS[6] = "Dolor de cabeza, diarrea o congestión nasal."
	QUESTIONS[7] = "Grupo vulnerable al Covid-19:\nLas personas consideradas dentro del grupo vulnerable son las que presenten alguna de las siguientes enfermedades o condición: Mayor de 65 años, asma severa o crónica, diabetes Mellitus, hipertensión arterial no controlada, enfermedades cardiovasculares graves, cáncer, enfermedad pulmonar crónica, insuficiencia renal en tratamiento con hemodiálisis, enfermedades o tratamientos de inmunosupresión u obesidad con índice de masa corporal mayor a 40.\n¿Pertenezco al grupo vulnerable al Covid-19?"
	QUESTIONS[8] = "¿Le falta la 3era dosis o dosis de refuerzo?"
	QUESTIONS[9] = "¿Cuenta con la 3era dosis o dosis de refuerzo?"

	var requests []ProcessRequest

	layout := "2006-01-02T15:04:05.000Z"
	date, err := time.Parse(layout, "2022-06-09T00:00:00.000Z")
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
			panic(fmt.Sprintf("Got error when try unmarshall string: %s", errUnmarshall))
		}
		for i := 0; i < 9; i++ {
			if i == 8 {
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
