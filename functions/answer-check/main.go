package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type Answer struct {
	IdQuestion string `json:"idQuestion"`
	Answer     bool   `json:"answer"`
}

type Question struct {
	Id         string `json:"id"`
	IdQuestion string `json:"sort"`
	Answer     bool   `json:"answer"`
	QQ         string `json:"question"`
	Order      int    `json:"order"`
}

type EventoPrev struct {
	Prev EventoResult `json:"prev"`
}

type EventoResult struct {
	Result Evento `json:"result"`
}

type Evento struct {
	Questions []Question `json:"qns"`
	Answers   []Answer   `json:"ans"`
}

func handler(ctx context.Context, ev EventoPrev) (string, error) {
	fmt.Println(ev)

	event := ev.Prev.Result
	var correcto []bool
	for _, quest := range event.Questions {
		for j, answ := range event.Answers {
			if quest.IdQuestion == answ.IdQuestion {
				if quest.Answer == answ.Answer {
					correcto = append(correcto, true)
				} else {
					correcto = append(correcto, false)
				}
				event.Answers[j] = event.Answers[len(event.Answers)-1]
				event.Answers = event.Answers[:len(event.Answers)-1]
				break
			}
		}
	}

	for _, resp := range correcto {
		if !resp {
			return "NO", nil
		}
	}
	return "SI", nil

}

func main() {
	lambda.Start(handler)
}
