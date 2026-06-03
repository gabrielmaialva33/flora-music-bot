package utils

import (
	"fmt"
	"time"

	"github.com/Laky-64/gologging"
	"resty.dev/v3"
)

const batbinBaseURL = "https://batbin.me/"

// Timeout: era o único client resty do projeto sem deadline — sem ele,
// CreatePaste (usado pra subir logs/erros) penduraria a goroutine se o batbin travasse.
var httpClient = resty.New().SetTimeout(30 * time.Second)

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
