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
	SignUp(email string, password string, name string) (string, error)
	AdminCreateUser(email string) (string error)
	AdminSetUserPassword(username string, password string) (string error)
	SignIn(email string, password string) (string error)
	ConfirmSignUp(username string, confirmationCode string) (string error)
	ResendConfirmationCode(username string) (string error)
	getUser(email string) ([]Response, error)
	ListUsers() ([]Response, error)
	AdminGetUser(username string) (string error)
	AdminDisableUser(username string) (string error)
	AdminEnableUser(username string) (string error)
	ChangePasswordUser(email string, password string, newPassword string) (string error)
	ForgotPassword(username string) (string error)
	ConfirmForgotPassword(email string, newPassword string, username string, confirmationCode string) (string error)
	ResendAdminCreateUser(email string, name string) (string error)
	UpdateUserAttributes(email string, password string, newEmail string, newName string) (string error)
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
	NewPassword      string `json:"newPassword"`
	Name             string `json:"name"`
	NewName          string `json:"newName"`
	ConfirmationCode string `json:"confirmationCode"`
	Case             int    `json:"case"`
	NewEmail         string `json:"newEmail"`
}

func (d *deps) handler(ctx context.Context, event Event) (string, error) {

	var result string
	var response []Response
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
	//s
	switch event.Case {
	case 0: // SignUp
		result, err = client.SignUp(event.Email, event.Password, event.Name)
	case 1: // AdminCreateUser
		result, err = client.AdminCreateUser(event.Email, event.Name)
	case 2: // AdminSetUserPassword
		result, err = client.AdminSetUserPassword(event.Username, event.Password)
	case 3: // SignIn
		result, err = client.SignIn(event.Email, event.Password)
	case 4: // ResendConfirmationCode
		result, err = client.ResendConfirmationCode(event.Username)
	case 5: // ConfirmSignUp
		result, err = client.ConfirmSignUp(event.Username, event.ConfirmationCode)
	case 6: // GetUser
		response, err = client.getUser(event.Email)
	case 7: // ListUsers
		response, err = client.ListUsers()
	case 8: // AdminGetUser
		result, err = client.AdminGetUser(event.Username)
	case 9: // AdminDisableUser
		result, err = client.AdminDisableUser(event.Username)
	case 10: // AdminEnableUser
		result, err = client.AdminEnableUser(event.Username)
	case 11: // ChangePassword
		result, err = client.ChangePasswordUser(event.Email, event.Password, event.NewPassword)
	case 12: // ForgotPassword
		result, err = client.ForgotPassword(event.Username)
	case 13: // ConfirmForgotPassword
		result, err = client.ConfirmForgotPassword(event.Email, event.NewPassword, event.Username, event.ConfirmationCode)
	case 14: // ResendAdminCreateUser
		result, err = client.ResendAdminCreateUser(event.Email, event.Name)
	case 15: // UpdateUserAttributes
		result, err = client.UpdateUserAttributes(event.Email, event.Password, event.NewEmail, event.NewName)
	}

	if err != nil {
		fmt.Println("Error :", err)
		return "", err
	}
	fmt.Println("CLIENTE :", client)
	fmt.Println("Response :", response)
	fmt.Println("result :", result)

	return "", nil
}

func main() {
	d := deps{}
	lambda.Start(d.handler)
}

func (ctx *awsCognitoClient) SignUp(email string, password string, name string) (string, error) {

	user := &cognito.SignUpInput{
		ClientId: aws.String(ctx.appClientId),
		Username: aws.String(email),
		Password: aws.String(password),
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
	fmt.Println("USER: ", user)

	result, err := ctx.cognitoClient.SignUp(user)
	if err != nil {
		fmt.Println("Error : SignUp", err)
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
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("true"),
			},
		},
	}
	fmt.Println("USER: ", user)

	result, err := ctx.cognitoClient.AdminCreateUser(user)
	if err != nil {
		fmt.Println("Error : AdminCreateUser", err)
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
		fmt.Println("Error : AdminSetUserPassword", err)
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

	fmt.Println("Resultado de InitiateAuth: ", result)

	if err != nil {
		fmt.Println("Error  : InitiateAuth", err)
		return "", err
	}

	return result.String(), nil
}

func (ctx *awsCognitoClient) ConfirmSignUp(username string, confirmationCode string) (string, error) {

	user := &cognito.ConfirmSignUpInput{
		ClientId:         aws.String(ctx.appClientId),
		ConfirmationCode: aws.String(confirmationCode),
		Username:         aws.String(username),
	}

	result, err := ctx.cognitoClient.ConfirmSignUp(user)
	if err != nil {
		fmt.Println("Error : ConfirmSignUp", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) ResendConfirmationCode(username string) (string, error) {

	user := &cognito.ResendConfirmationCodeInput{
		ClientId: aws.String(ctx.appClientId),
		Username: aws.String(username),
	}

	result, err := ctx.cognitoClient.ResendConfirmationCode(user)
	if err != nil {
		fmt.Println("Error : ResendConfirmationCode", err)
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
		fmt.Println("Got error ListUsers in function getUser")
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
		fmt.Println("Error : AdminGetUser")
		os.Exit(1)
	}

	fmt.Println(result)

	return result.String(), nil
}

func (ctx *awsCognitoClient) AdminDisableUser(username string) (string, error) {

	adminDisableUserInput := &cognito.AdminDisableUserInput{
		UserPoolId: aws.String(ctx.userPoolId),
		Username:   aws.String(username),
	}

	result, err := ctx.cognitoClient.AdminDisableUser(adminDisableUserInput)

	if err != nil {
		fmt.Println("Error: AdminDisableUser")
		os.Exit(1)
	}

	return result.String(), nil
}

func (ctx *awsCognitoClient) AdminEnableUser(username string) (string, error) {

	adminEnableUserInput := &cognito.AdminEnableUserInput{
		UserPoolId: aws.String(ctx.userPoolId),
		Username:   aws.String(username),
	}

	result, err := ctx.cognitoClient.AdminEnableUser(adminEnableUserInput)

	if err != nil {
		fmt.Println("Error: AdminEnableUser")
		os.Exit(1)
	}

	return result.String(), nil
}

func (ctx *awsCognitoClient) ChangePasswordUser(email string, password string, newPassword string) (string, error) {

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

	fmt.Println(result)

	accessToken := result.AuthenticationResult.AccessToken

	fmt.Println("Token expires in ", result.AuthenticationResult.ExpiresIn)
	fmt.Println("AccessToken: ", accessToken)

	changePasswordInput := &cognito.ChangePasswordInput{
		AccessToken:      aws.String(*result.AuthenticationResult.AccessToken),
		PreviousPassword: aws.String(password),
		ProposedPassword: aws.String(newPassword),
	}

	result2, err2 := ctx.cognitoClient.ChangePassword(changePasswordInput)

	if err2 != nil {
		fmt.Println("Error  : ChangePassword", err2)
		return "", err2
	}

	return result2.String(), nil
}

func (ctx *awsCognitoClient) ForgotPassword(username string) (string, error) {

	forgotPasswordInput := &cognito.ForgotPasswordInput{
		ClientId: aws.String(ctx.appClientId),
		Username: aws.String(username),
	}

	result2, err2 := ctx.cognitoClient.ForgotPassword(forgotPasswordInput)

	if err2 != nil {
		fmt.Println("Error  : ForgotPassword", err2)
		return "", err2
	}

	println(result2.CodeDeliveryDetails.DeliveryMedium)

	return result2.String(), nil
}

func (ctx *awsCognitoClient) ConfirmForgotPassword(email string, newPassword string, username string, confirmationCode string) (string, error) {

	confirmForgotPasswordInput := &cognito.ConfirmForgotPasswordInput{
		ClientId:         aws.String(ctx.appClientId),
		Username:         aws.String(username),
		ConfirmationCode: aws.String(confirmationCode),
		Password:         aws.String(newPassword),
	}

	result2, err2 := ctx.cognitoClient.ConfirmForgotPassword(confirmForgotPasswordInput)

	if err2 != nil {
		fmt.Println("Error  : ConfirmForgotPassword", err2)
		return "", err2
	}

	return result2.String(), nil
}

func (ctx *awsCognitoClient) ResendAdminCreateUser(email string, name string) (string, error) {

	user := &cognito.AdminCreateUserInput{
		UserPoolId:    aws.String(ctx.userPoolId),
		Username:      aws.String(email),
		MessageAction: aws.String("RESEND"),
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
	fmt.Println("USER: ", user)

	result, err := ctx.cognitoClient.AdminCreateUser(user)
	if err != nil {
		fmt.Println("Error : AdminCreateUser", err)
		return "", err
	}
	return result.String(), nil
}

func (ctx *awsCognitoClient) UpdateUserAttributes(email string, password string, newEmail string, newName string) (string, error) {

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

	fmt.Println(result)

	updateUserAttributesInput := &cognito.UpdateUserAttributesInput{
		AccessToken: aws.String(*result.AuthenticationResult.AccessToken),
		UserAttributes: []*cognito.AttributeType{
			{
				Name:  aws.String("name"),
				Value: aws.String(newName),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(newEmail),
			},
		},
	}
	fmt.Println("Error 111")

	result2, err2 := ctx.cognitoClient.UpdateUserAttributes(updateUserAttributesInput)

	fmt.Println(result2)

	if err2 != nil {
		fmt.Println("Error  : UpdateUserAttributes", err2)
		return "", err2
	}

	return result2.String(), nil

}
