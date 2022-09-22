package main

import (
	"bytes"
	"context"
	"fmt"

	"html/template"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// const (

// 	// HTMLBody ...
// 	HTMLBody = "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
// 		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
// 		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>.</p>"

// 	//The email body for recipients with non-HTML email clients.

// 	// The character encoding for the email.
// 	charSet = "UTF-8"
// )

type UserAttributesType struct {
	CognitoUser       string `json:"cognito:user_status"`
	Nombre            string `json:"name"`
	Email             string `json:"email"`
	Sub               string `json:"sub"`
	UsernameParameter string `json:"usernameParameter"`
}

type CognitoEventUserPoolsCallerContext struct {
	AWSSDKVersion string `json:"awsSdkVersion"`
	ClientID      string `json:"clientId"`
}

type CognitoEventUserPoolsHeader struct {
	Version       string                             `json:"version"`
	TriggerSource string                             `json:"triggerSource"`
	Region        string                             `json:"region"`
	UserPoolID    string                             `json:"userPoolId"`
	CallerContext CognitoEventUserPoolsCallerContext `json:"callerContext"`
	UserName      string                             `json:"userName"`
}

// CognitoEventUserPoolsCustomMessage is sent by AWS Cognito User Pools before a verification or MFA message is sent,
// allowing a user to customize the message dynamically.
type CognitoEventUserPoolsCustomMessage struct {
	CognitoEventUserPoolsHeader
	Request  CognitoEventUserPoolsCustomMessageRequest  `json:"request"`
	Response CognitoEventUserPoolsCustomMessageResponse `json:"response"`
}

// CognitoEventUserPoolsCustomMessageRequest contains the request portion of a CustomMessage event
type CognitoEventUserPoolsCustomMessageRequest struct {
	UserAttributes    UserAttributesType `json:"userAttributes"`
	CodeParameter     string             `json:"codeParameter"`
	UsernameParameter string             `json:"usernameParameter"`
	ClientMetadata    map[string]string  `json:"clientMetadata"`
}

// CognitoEventUserPoolsCustomMessageResponse contains the response portion of a CustomMessage event
type CognitoEventUserPoolsCustomMessageResponse struct {
	SMSMessage   string `json:"smsMessage"`
	EmailMessage string `json:"emailMessage"`
	EmailSubject string `json:"emailSubject"`
}

func handler(ctx context.Context, event CognitoEventUserPoolsCustomMessage) (CognitoEventUserPoolsCustomMessage, error) {
	fmt.Println("Evento: ", event)
	// CONECTAR SESSION CON AWS
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*aws.String("us-east-1"))},
	)
	if err != nil {
		fmt.Println(err.Error())
		return CognitoEventUserPoolsCustomMessage{}, err
	}
	var bodyHtml string
	switch event.TriggerSource {
		case "CustomMessage_SignUp":
			bodyHtml = "signUpMail"
		case "CustomMessage_ResendCode":
			bodyHtml = "resendCodeMail"
		case "CustomMessage_ForgotPassword":
			bodyHtml = "forgotPasswordMail"
		case "CustomMessage_AdminCreateUser":
			bodyHtml = "adminCreateUserMail"
	}

	// CONECTAR SESSION CON S3
	svc := s3.New(sess)
	// OBTENER EL TEMPLATE HTML
	rawObject, err := svc.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv("BucketName")),
			Key:    aws.String(fmt.Sprintf("%s.html", bodyHtml)),
		})
	if err != nil {
		fmt.Println(err.Error())
		return CognitoEventUserPoolsCustomMessage{}, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(rawObject.Body)
	HTMLBody := buf.String()
	// CONSTRUIR EL HTML
	t, err := template.New("mailhtml").Parse(HTMLBody)
	if err != nil {
		fmt.Println(err.Error())
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, event.Request.UserAttributes); err != nil {
		fmt.Println(err.Error())
	}

	resultTemplate := tpl.String()

	//fmt.Println("Email Sent to address: " + event.Request.UserAttributes.Email)

	event.Response.EmailMessage = resultTemplate
	event.Response.EmailSubject = "Invitación a Oficina Virtual: Cuenta Nueva"

	return event, nil

}

func main() {
	lambda.Start(handler)
}
