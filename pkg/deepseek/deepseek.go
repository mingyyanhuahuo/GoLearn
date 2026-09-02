package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func Enabled() bool {
	return apiKey != ""
}

var (
	apiKey  string
	baseURL = "https://api.deepseek.com"
	model   = "deepseek-v4-flash"
)

func Init(key, url, mdl string) {
	apiKey = key
	if url != "" {
		baseURL = url
	}
	if mdl != "" {
		model = mdl
	}
}
func Chat(messages []Message) (string, error) {
	if !Enabled() {
		return "", errors.New("DeepSeek API 未配置")
	}
	body, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, baseURL+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API 请求失败: %d %s", resp.StatusCode, string(data))
	}
	var result chatResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", errors.New("DeepSeek 返回空结果")
	}
	return result.Choices[0].Message.Content, nil
}
