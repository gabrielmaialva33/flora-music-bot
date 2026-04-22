package utils

import (
	"fmt"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"
)

const batbinBaseURL = "https://batbin.me/"

var httpClient = resty.New()

type batbinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func CreatePaste(content string) (string, error) {
	var result batbinResponse

	resp, err := httpClient.R().
		SetBody(content).
		SetResult(&result).
		Post(batbinBaseURL + "api/v2/paste")
	if err != nil {
		gologging.Error("batbin request error: " + err.Error())
		return "", err
	}

	if resp.StatusCode() != 200 {
		gologging.Error("batbin bad response: " + resp.String())
		return "", fmt.Errorf("batbin returned status %d", resp.StatusCode())
	}

	if !result.Success {
		err := fmt.Errorf("batbin paste failed")
		gologging.Error(err.Error())
		return "", err
	}

	return batbinBaseURL + result.Message, nil
}
