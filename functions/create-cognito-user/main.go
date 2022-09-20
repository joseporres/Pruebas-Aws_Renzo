package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type deps struct {
}

type CognitoClient interface {
	SignUp(email string, password string) (string, error)
	AdminCreateUser(email string) (string, error)
	AdminSetUserPassword(username string, password string) (string error)
	SignIn(email string, password string) (error, string)
}

type awsCognitoClient struct {
	cognitoClient *cognito.CognitoIdentityProvider
	appClientId   string
	userPoolId    string
}

type Event struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Case     int    `json:"case"`
}

func (d *deps) handler(ctx context.Context, event Event) (string, error) {
	// CONECTAR SESSION CON AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*aws.String("us-east-1"))},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to connect session, %v", err))
	}
	// INICIAR SESSION EN COGNITO
	svc := cognito.New(sess)

	fmt.Println("APP_CLIENT_ID : ", os.Getenv("APP_CLIENT_ID"))
	fmt.Println("USER_POOL_ID : ", os.Getenv("USER_POOL_ID"))

	client := awsCognitoClient{
		cognitoClient: svc,
		appClientId:   os.Getenv("APP_CLIENT_ID"),
		userPoolId:    os.Getenv("USER_POOL_ID"),
	}
	fmt.Printf("Email :%s Password: %s \n", event.Email, event.Password)
	fmt.Println("cliente: ", client)

	switch event.Case {
	case 0: // SignUp
		client.SignUp(event.Email, event.Password)
	case 1: // AdminCreateUser
		client.AdminCreateUser(event.Email, event.Name)
	case 2: // AdminSetUserPassword
		client.AdminSetUserPassword(event.Username, event.Password)
	case 3: // SignIn
		client.SignIn(event.Email, event.Password)
	}

	fmt.Print(client)
	return "", nil
}

func main() {
	d := deps{}
	lambda.Start(d.handler)
}

func (ctx *awsCognitoClient) SignUp(email string, password string) (string, error) {

	user := &cognito.SignUpInput{
		ClientId: aws.String(ctx.appClientId),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
		},
	}
	fmt.Println("USER: ", user)

	result, err := ctx.cognitoClient.SignUp(user)
	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) AdminCreateUser(email string, name string) (string, error) {

	user := &cognito.AdminCreateUserInput{
		UserPoolId: aws.String(ctx.userPoolId),
		Username:   aws.String(email),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(name),
			},
		},
	}
	fmt.Println("USER: aaaa ", user)

	result, err := ctx.cognitoClient.AdminCreateUser(user)
	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) AdminSetUserPassword(username string, password string) (string, error) {

	user := &cognito.AdminSetUserPasswordInput{
		UserPoolId: aws.String(ctx.userPoolId),
		Username:   aws.String(username),
		Password:   aws.String(password),
		Permanent:  aws.Bool(true),
	}

	result, err := ctx.cognitoClient.AdminSetUserPassword(user)
	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}

	return result.String(), nil
}

func (ctx *awsCognitoClient) SignIn(email string, password string) (string, error) {
	initiateAuthInput := &cognito.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: aws.StringMap(map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		}),
		ClientId: aws.String(ctx.appClientId),
	}

	result, err := ctx.cognitoClient.InitiateAuth(initiateAuthInput)

	if err != nil {
		fmt.Println("Error  : InitiateAuth", err)
		return "", err
	}

	return result.String(), nil
}
