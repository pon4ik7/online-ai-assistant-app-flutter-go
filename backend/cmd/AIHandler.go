package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// UserMessage - a structure to save the user query and send to the DeepSeek
type UserMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// The AIRequest DeepSeek would handle
type AIRequest struct {
	Model    string        `json:"model"`
	Messages []UserMessage `json:"messages"`
}

// The Response DeepSeek returns
type Response struct {
	Choices []struct {
		ResponseContent struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func HandleUserQuery(messagesCache map[string]map[string]string, query string, isAdmin bool, sessionID string) (string, error) {
	initialPrompt := query
	query = ClarifyProductContext(messagesCache, sessionID) + query
	response, err := GetResponse(query, isAdmin)
	if err != nil {
		log.Println(err)
		return "", err
	} else {
		cacheMessage(messagesCache, initialPrompt, response, sessionID)
	}
	return response, nil
}

func ClarifyProductContext(messagesCache map[string]map[string]string, sessionID string) string {
	//TODO extract data from the data base and offer to the DeepSeek as well

	instructions := "You are a friendly store consultant helping a customer. Follow these rules: \n" +
		"Always address the customer with formal \"Вы\" (Russian) or \"you\" in a respectful tone (English)\n" +
		"Respond in the same language as the QUESTION:\n " +
		"Use professional but warm language:  \n " +
		"Avoid robotic phrases (\"Based on your query...\")\n" +
		"Treat the CONTEXT: as our prior conversation history\n" +
		"Acknowledge past discussions naturally\n" +
		"Reference prior interactions if appropriate\n" +
		"For complex queries, offer step-by-step guidance (I recommend checking the size first, then I’ll assist with payment)\n" +
		"Use polite fillers\n" +
		"Prioritize clarity and empathy\n" +
		"Answer briefly but concise and meaningful\n" +
		"Do not answer the questions that are not asked\n" +
		"Greet the customer only once do not use \"Здравствуйте\" and Hello each message\n" +
		"If the QUESTION: is unclear, ask for details like a human would\n CONTEXT:"
	for query, response := range messagesCache[sessionID] {
		concat := query + ":" + response
		instructions += "\n" + concat
	}

	return instructions + "QUESTION: "
}

func GetResponse(query string, isAdmin bool) (string, error) {
	var role string
	if isAdmin {
		role = "system"
	} else {
		role = "user"
	}

	request := AIRequest{
		Model: "deepseek-chat", //deepseek-reasoner for complex topics
		Messages: []UserMessage{
			{
				Role:    role,
				Content: query,
			},
		},
	}

	jsonEncoded, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to encode request: %v", err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewBuffer(jsonEncoded),
	)

	if err != nil {
		return "", fmt.Errorf("failed to create html request: %v", err)
	}

	envPath := filepath.Join(".", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Не удалось загрузить .env файл: %v", err)

		// Проверяем переменные окружения напрямую
		apiKey := os.Getenv("API_KEY")
		if apiKey == "" {
			log.Fatal("API_KEY не найден ни в .env файле, ни в переменных окружения")
		}
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("API_KEY не установлен")
	}
	//apiKey := "sk-9b414ce298bf4fe2b481cdd129de5be0"

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	response, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("API request failed: %v", err)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API returned status: %d", response.StatusCode)
	}

	var aiReply Response
	if err := json.NewDecoder(response.Body).Decode(&aiReply); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(aiReply.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return aiReply.Choices[0].ResponseContent.Content, nil
}

func cacheMessage(messagesCache map[string]map[string]string, query string, response string, sessionID string) {
	if _, ok := messagesCache[sessionID]; !ok {
		messagesCache[sessionID] = make(map[string]string)
	}
	messagesCache[sessionID][query] = response
}

func SaveDialogueContext(keyWords string, err error) {
	if err != nil {
		log.Fatalf("Failed to save dialogue context: %v", err)
		return
	}
	fmt.Println(keyWords)
	//TODO save the keywords into the DB for 15 minutes unless the client comes back
}
