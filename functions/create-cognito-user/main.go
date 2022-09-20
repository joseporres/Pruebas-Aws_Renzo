package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cognito "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type deps struct {
}

type CognitoClient interface {
	SignUp(email string, password string) (string, error)
	AdminCreateUser(email string) (string error)
	AdminSetUserPassword(username string, password string) (string error)
	SignIn(email string, password string) (string error)
	ConfirmSignUp(email string, username string, confirmationCode string) (string error)
	ResendConfirmationCode(email string, username string) (string error)
	getUser(email string) ([]Response, error)
	ListUsers() ([]Response, error)
	AdminGetUser(username string) (string error)
}

type awsCognitoClient struct {
	cognitoClient *cognito.CognitoIdentityProvider
	appClientId   string
	userPoolId    string
}

type Response struct {
	Username            string    `json:"username"`
	Enabled             bool      `json:"enabled"`
	AccountStatus       string    `json:"accountStatus"`
	Email               string    `json:"email"`
	EmailVerified       string    `json:"emailVerified"`
	PhoneNumberVerified string    `json:"phoneNumberVerified"`
	Updated             time.Time `json:"updated"`
	Created             time.Time `json:"created"`
}

type Event struct {
	Email            string `json:"email"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	ConfirmationCode string `json:"confirmationCode"`
	Case             int    `json:"case"`
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
	case 4: // ResendConfirmationCode
		client.ResendConfirmationCode(event.Email, event.Password)
	case 5: // ConfirmSignUp
		client.ConfirmSignUp(event.Email, event.Username, event.ConfirmationCode)
	case 6: // GetUser
		client.getUser(event.Email)
	case 7: // ListUsers
		client.ListUsers()
	case 8: // AdminGetUser
		client.AdminGetUser(event.Username)
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

func (ctx *awsCognitoClient) ConfirmSignUp(email string, username string, confirmationCode string) (string, error) {

	user := &cognito.ConfirmSignUpInput{
		ClientId:         aws.String(ctx.appClientId),
		ConfirmationCode: aws.String(confirmationCode),
		Username:         aws.String(username),
	}
	fmt.Println("USER: aaaa ", user.Username)

	result, err := ctx.cognitoClient.ConfirmSignUp(user)
	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) ResendConfirmationCode(email string, username string) (string, error) {

	user := &cognito.ResendConfirmationCodeInput{
		ClientId: aws.String(ctx.appClientId),
		Username: aws.String(username),
	}
	fmt.Println("USER: aaaa ", user.Username)

	result, err := ctx.cognitoClient.ResendConfirmationCode(user)
	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) getUser(email string) ([]Response, error) {

	user := &cognito.ListUsersInput{
		Filter:     aws.String("email = \"" + email + "\""),
		UserPoolId: aws.String(ctx.userPoolId),
	}

	result, err := ctx.cognitoClient.ListUsers(user)

	if err != nil {
		fmt.Println("Got error listing users")
		os.Exit(1)
	}

	var response []Response
	for _, user := range result.Users {
		fmt.Println("user: ", user)
		response = append(response, Response{
			Username:      *user.Username,
			Enabled:       *user.Enabled,
			AccountStatus: *user.UserStatus,
			Email:         *user.Attributes[0].Value,
			Updated:       *user.UserLastModifiedDate,
			Created:       *user.UserCreateDate,
		})
	}

	return response, nil

}

func (ctx *awsCognitoClient) ListUsers() ([]Response, error) {

	user := &cognito.ListUsersInput{
		UserPoolId: aws.String(ctx.userPoolId),
	}

	result, err := ctx.cognitoClient.ListUsers(user)

	if err != nil {
		fmt.Println("Got error listing users")
		os.Exit(1)
	}

	var response []Response
	for _, user := range result.Users {
		fmt.Println("user: ", user)
		response = append(response, Response{
			Username:      *user.Username,
			Enabled:       *user.Enabled,
			AccountStatus: *user.UserStatus,
			Updated:       *user.UserLastModifiedDate,
			Created:       *user.UserCreateDate,
		})
	}

	return response, nil
}

func (ctx *awsCognitoClient) AdminGetUser(username string) (string, error) {

	user := &cognito.AdminGetUserInput{
		UserPoolId: aws.String(ctx.userPoolId),
		Username:   aws.String(username),
	}

	result, err := ctx.cognitoClient.AdminGetUser(user)

	if err != nil {
		fmt.Println("Got error listing users")
		os.Exit(1)
	}

	fmt.Println(result)

	return result.String(), nil

}
