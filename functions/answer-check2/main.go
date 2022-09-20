package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoString struct {
	S string `json:"S"`
}

type DynamoBool struct {
	Bool bool `json:"Bool"`
}

type Question struct {
	Answer     DynamoBool   `json:"answer"`
	IdQuestion DynamoString `json:"sort"`
	Question   DynamoString `json:"question"`
}

type UserAnswer struct {
	IdQuestion string `json:"idQuestion"`
	Answer     bool   `json:"answer"`
}

type Result struct {
	State       int                      `json:"state"`
	QuizAnswers *dynamodb.AttributeValue `json:"quizAnswers"`
}

type QuizAnswer struct {
	Question string `json:"question"`
	Answer   bool   `json:"answer"`
}

type Event struct {
	UserAnswers []UserAnswer `json:"userAnswers"`
	Questions   []Question   `json:"questions"`
}

func handler(ctx context.Context, ev Event) (Result, error) {

	var answers []QuizAnswer

	for _, ques := range ev.Questions {
		for j, answ := range ev.UserAnswers {
			if ques.IdQuestion.S == answ.IdQuestion {
				if ques.Answer.Bool != answ.Answer {
					// retorna estado 6 y array de respuestas vacio
					return Result{6, nil}, nil
				}
				answers = append(answers, QuizAnswer{ques.Question.S, answ.Answer})
				// Esto es para acortar la cantidad de loops
				ev.UserAnswers[j] = ev.UserAnswers[len(ev.UserAnswers)-1]
				ev.UserAnswers = ev.UserAnswers[:len(ev.UserAnswers)-1]
				break
			}
		}
	}
	dbAnswers, err := dynamodbattribute.Marshal(answers)
	if err != nil {
		fmt.Println(err.Error())
		return Result{6, nil}, nil
	}
	// retorna estado 5 y array de respuestas
	return Result{5, dbAnswers}, nil

}

func main() {
	lambda.Start(handler)
}
