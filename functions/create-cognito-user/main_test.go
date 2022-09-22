package main

import (
	"context"
	"testing"
)

func TestHandler(t *testing.T) {
	t.Run("success request", func(t *testing.T) {
		d := deps{}
		k, err := d.handler(context.TODO(), Event{Email: "japorres18@gmail.com", Username: "id", Password: "PaSsWoRd_100", Name: "Jose 	", ConfirmationCode: "11", Case: 1})
		if err != nil {
			t.Fatal("ErroraaaaaaaaaaaaaaaaaaAaa")
		}
		if k != "" {
			t.Fatal("Error")
		}
	})
}
