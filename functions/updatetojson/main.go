package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

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
	Id                  string   `json:"id"`
	Sort                string   `json:"sort"`
	Name                string   `json:"name"`
	DocType             string   `json:"docType"`
	Dni                 string   `json:"dni"`
	Gender              string   `json:"gender"`
	BirthDate           string   `json:"birthDate"`
	CountryOfBirth      string   `json:"countryOfBirth"`
	PersonalEmail       string   `json:"personalEmail"`
	MaritalStatus       string   `json:"maritalStatus"`
	PersonalPhone       string   `json:"personalPhone"`
	CountryOfResidence  string   `json:"countryOfResidence"`
	ResidenceDepartment string   `json:"residenceDepartment"`
	Address             string   `json:"address"`
	Area                string   `json:"area"`
	SubArea             string   `json:"subArea"`
	WorkerType          string   `json:"workerType"`
	Email               string   `json:"email"`
	CreationDate        string   `json:"creationDate"`
	UpdateLastSession   string   `json:"updateLastSession"`
	EntryDate           string   `json:"entryDate"`
	Phone               string   `json:"phone"`
	Apps                []Option `json:"apps"`
	Menu                []Option `json:"menu"`
	Processes           []Option `json:"processes"`
	UserType            string   `json:"userType"`
	UserStatus          string   `json:"userStatus"`
	Role                string   `json:"role"`
	Days                int      `json:"days"`
	HomeOffice          int      `json:"homeOffice"`
	Photo               string   `json:"photo"`
	Boss                string   `json:"boss"`
	BossName            string   `json:"bossName"`
	User                string   `json:"user"`
	Backup              string   `json:"backup"`
	BackupName          string   `json:"backupName"`
	Access              bool     `json:"access"`
}

type Event struct {
	UserId string `json:"userId"`
}

func handler(ctx context.Context, event Event) (UserObject, error) {

	TABLE_NAME := os.Getenv("TABLA_NAME")
	SORT := "SETTINGS"

	timeNow := time.Now()
	utcPeru := timeNow.Add(-5 * time.Hour)
	now := strings.Split(utcPeru.Format(time.RFC3339), "T")[0]

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return UserObject{}, err
	}

	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(TABLE_NAME),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(event.UserId),
			},
			"sort": {
				S: aws.String(SORT),
			},
		},
		UpdateExpression: aws.String("set updateLastSession = :updateLastSession"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":updateLastSession": {
				S: aws.String(now),
			},
		},
		ReturnValues: aws.String("ALL_NEW"),
	}
	result, err := svc.UpdateItem(input)
	if err != nil {
		panic(fmt.Sprintf("failed to Dynamodb Update Items, %v", err))
	}

	user := UserObject{}

	fmt.Println(result.Attributes)

	dynamodbattribute.UnmarshalMap(result.Attributes, &user)

	fmt.Println(user)

	user.Access = true

	return user, nil
}

func main() {
	lambda.Start(handler)
}
