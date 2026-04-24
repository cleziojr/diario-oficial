package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var APIURL = "https://openrouter.ai/api/v1/chat/completions"

type HFRequest struct {
	Model    string      `json:"model"`
	Messages []HFMessage `json:"messages"`
}

type HFMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type HFResponse struct {
	Choices []struct {
		Message HFMessage `json:"message"`
	} `json:"choices"`
}

func Summarize(text string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY não definida")
	}

	reqBody := HFRequest{
		Model: "google/gemma-3-12b-it:free",
		Messages: []HFMessage{
			{
				Role:    "user",
				Content: "Resuma o texto em tópicos claros:\n\n" + text,
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/cleziojr/diario-oficial")
	req.Header.Set("X-Title", "Insight Diário")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("Resposta crua:", string(bodyBytes))

	var result HFResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia")
	}

	return result.Choices[0].Message.Content, nil
}
