package main

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIClient struct {
	client openai.Client
	model  string
}

func NewOpenAIClient(apiKey, baseURL, model string) *OpenAIClient {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	client := openai.NewClient(opts...)

	if model == "" {
		model = "gpt-4o"
	}

	return &OpenAIClient{
		client: client,
		model:  model,
	}
}

func (o *OpenAIClient) CreateReview(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(userPrompt),
	}

	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    o.model,
	}

	completion, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("getting chat completion: %w", err)
	}

	return completion.Choices[0].Message.Content, nil
}
