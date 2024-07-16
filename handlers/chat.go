package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleChat(c *gin.Context) {
	message := strings.TrimSpace(c.PostForm("message"))
	if message == "" {
		c.String(http.StatusBadRequest, "Message cannot be empty")
		return
	}

	response, err := callHuggingFaceAPI(message)
	if err != nil {
		log.Printf("Error calling Hugging Face API: %v", err)
		c.String(http.StatusInternalServerError, "Error processing request")
		return
	}

	// Return only the bot's response as plain text
	c.String(http.StatusOK, response)
}
func callHuggingFaceAPI(message string) (string, error) {
	url := os.Getenv("MODELURL")
	apiKey := os.Getenv("HUGGINGFACE_API_KEY")

	if url == "" || apiKey == "" {
		return "", fmt.Errorf("MODELURL or HUGGINGFACE_API_KEY environment variable is not set")
	}

	payload := map[string]interface{}{
		"inputs": fmt.Sprintf("Human: %s\nAI:", message),
		"parameters": map[string]interface{}{
			"max_new_tokens": 150,
			"temperature":    0.7,
			"top_p":          0.95,
			"do_sample":      true,
			"stop":           []string{"\nHuman:", "\n\nHuman:"},
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	if len(result) > 0 {
		if generatedText, ok := result[0]["generated_text"].(string); ok {
			// Extract only the AI's response
			parts := strings.Split(generatedText, "AI:")
			if len(parts) > 1 {
				aiResponse := strings.TrimSpace(parts[1])
				// Remove any remaining "Human:" parts
				aiResponse = strings.Split(aiResponse, "Human:")[0]
				return strings.TrimSpace(aiResponse), nil
			}
			return "", fmt.Errorf("unexpected response format: AI response not found")
		}
		return "", fmt.Errorf("unexpected response format")
	}

	return "", fmt.Errorf("empty response from API")
}

