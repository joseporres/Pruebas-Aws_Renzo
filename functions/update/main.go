package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type PutObject struct {
	Id       string `json:"id"`
	Sort     string `json:"sort"`
	Name     string `json:"name"`
	Cadena   string `json:"putCadena"`
	Entero   int    `json:"putEntero"`
	Booleano bool   `json:"putBooleano"`
	Vacio    string `json:"putVacio"`
}

type Data struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Cadena   string `json:"cadena"`
	Entero   int    `json:"entero"`
	Booleano bool   `json:"booleano"`
}

type Event struct {
	DataObject Data   `json:"data"`
	Vacio      string `json:"vacio"`
}

func handler(ctx context.Context, event Event) (string, error) {
	TABLE_NAME := os.Getenv("TABLA_NAME")

	fmt.Print("event : ", event, "\n")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return "", err
	}

	svc := dynamodb.New(sess)

	enteroStr := strconv.Itoa(event.DataObject.Entero)

	putObject := PutObject{
		Id:       "TEST",
		Sort:     event.DataObject.Id,
		Name:     event.DataObject.Name,
		Cadena:   event.DataObject.Cadena,
		Entero:   event.DataObject.Entero,
		Booleano: event.DataObject.Booleano,
		Vacio:    event.Vacio,
	}

	fmt.Print("putObject : ", putObject, "\n")

	putItem, err := dynamodbattribute.MarshalMap(putObject)
	if err != nil {
		return "", err
	}

	fmt.Print("putItem : ", putItem, "\n")

	putInput := &dynamodb.PutItemInput{
		Item:      putItem,
		TableName: aws.String(TABLE_NAME),
	}

	fmt.Print("putInput   : ", putInput, "\n")

	_, err = svc.PutItem(putInput)
	if err != nil {
		return "", err
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String("TEST"),
			},
			"sort": {
				S: aws.String(event.DataObject.Id),
			},
		},
		UpdateExpression: aws.String("set #name = :name, updateCadena = :updateCadena, updateEntero = :updateEntero, updateBooleano = :updateBooleano, updateVacio = :updateVacio"),
		ExpressionAttributeNames: map[string]*string{
			"#name": aws.String("name"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":name": {
				S: aws.String(event.DataObject.Name),
			},
			":updateCadena": {
				S: aws.String(event.DataObject.Cadena),
			},
			":updateEntero": {
				N: aws.String(enteroStr),
			},
			":updateBooleano": {
				BOOL: aws.Bool(event.DataObject.Booleano),
			},
			":updateVacio": {
				S: aws.String(event.Vacio),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
	}
	_, err1 := svc.UpdateItem(updateInput)
	if err1 != nil {
		panic(fmt.Sprintf("failed to Dynamodb Update Items, %v", err))
	}

	return "Success", nil
}

func main() {
	lambda.Start(handler)
}

func MarshalMap(in interface{}) (map[string]*dynamodb.AttributeValue, error) {
	av, err := getEncoder().Encode(in)
	if err != nil || av == nil || av.M == nil {
		return map[string]*dynamodb.AttributeValue{}, err
	}

	return av.M, nil
}

func getEncoder() *dynamodbattribute.Encoder {
	encoder := dynamodbattribute.NewEncoder()
	encoder.NullEmptyString = false
	return encoder
}
