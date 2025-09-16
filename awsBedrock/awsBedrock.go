package awsbedrock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

const MODEL = "us.meta.llama4-maverick-17b-instruct-v1:0"

// us.meta.llama4-maverick-17b-instruct-v1:0
// "us.meta.llama4-scout-17b-instruct-v1:0"
// us.meta.llama3-1-70b-instruct-v1:0
// us.anthropic.claude-3-5-haiku-20241022-v1:0

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type SCHEMA string

func New() *BedrockClient {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		return nil
	}

	client := bedrockruntime.NewFromConfig(sdkConfig)
	return &BedrockClient{
		BedrockRuntimeClient: client,
	}

}

type BedrockClient struct {
	BedrockRuntimeClient *bedrockruntime.Client
	Request              any
	System               string
	PastMessages         []Message
	Schema               string
}

type Embedding struct {
	InputText string `json:"inputText"`
}

type EmbeddingResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

func removeJsonAItags(content string) string {
	content = strings.ReplaceAll(content, "```json", "")
	content = strings.ReplaceAll(content, "```", "")
	return content
}

func (wrapper BedrockClient) GenerateAI(request string, system string, pastMessages []map[string]string) (string, error) {
	modelId := MODEL
	ctx := context.TODO()
	var contentBlocks []types.Message

	// add past messages
	for _, msg := range pastMessages {
		contentBlocks = append(contentBlocks, types.Message{
			Role: types.ConversationRole(msg["role"]),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: fmt.Sprintf("%v", msg["content"]),
				},
			},
		})
	}

	contentBlocks = append(contentBlocks, types.Message{
		Role: types.ConversationRole("user"),
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{
				Value: fmt.Sprintf("%v", request),
			},
		},
	})

	output, err := wrapper.BedrockRuntimeClient.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId:  aws.String(modelId),
		Messages: contentBlocks,
		System: []types.SystemContentBlock{&types.SystemContentBlockMemberText{
			Value: system,
		}},
	})

	if err != nil {
		return "", fmt.Errorf("model err: %s : %w", modelId, err)
	}

	responseText, _ := output.Output.(*types.ConverseOutputMemberMessage)
	responseContentBlock := responseText.Value.Content[0]
	text, _ := responseContentBlock.(*types.ContentBlockMemberText)
	final := removeJsonAItags(text.Value)
	return final, nil
}

func (wrapper BedrockClient) AnthropicAI(request any, system string, pastMessages []Message) (string, error) {
	modelId := "us.anthropic.claude-3-5-haiku-20241022-v1:0"
	ctx := context.TODO()
	var contentBlocks []types.Message

	// add past messages
	for _, msg := range pastMessages {
		contentBlocks = append(contentBlocks, types.Message{
			Role: types.ConversationRole(msg.Role),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: fmt.Sprintf("%v", msg.Content),
				},
			},
		})
	}

	contentBlocks = append(contentBlocks, types.Message{
		Role: types.ConversationRole("user"),
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{
				Value: fmt.Sprintf("%v", request),
			},
		},
	})

	output, err := wrapper.BedrockRuntimeClient.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId:  aws.String(modelId),
		Messages: contentBlocks,
		System: []types.SystemContentBlock{&types.SystemContentBlockMemberText{
			Value: system,
		}},
	})

	if err != nil {
		return "", fmt.Errorf("model err: %s : %w", modelId, err)
	}

	responseText, _ := output.Output.(*types.ConverseOutputMemberMessage)
	responseContentBlock := responseText.Value.Content[0]
	text, _ := responseContentBlock.(*types.ContentBlockMemberText)
	final := removeJsonAItags(text.Value)
	return final, nil
}

func (wrapper BedrockClient) GetEmbeddings(prompt string) ([]float32, error) {
	// modelId := "amazon.titan-embed-text-v2:0"
	modelId := "amazon.titan-embed-text-v1"
	ctx := context.TODO()
	body, err := json.Marshal(Embedding{
		InputText: prompt,
	})

	if err != nil {
		log.Fatal("failed to marshal", err)
	}

	output, err := wrapper.BedrockRuntimeClient.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelId),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil {
		return nil, fmt.Errorf("model err: %s : %w", modelId, err)
	}

	var response EmbeddingResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		log.Fatal("failed to unmarshal", err)
	}

	return response.Embedding, nil
}

func (wrapper BedrockClient) AnthropicAISchema(request any, system string, pastMessages []Message, schema SCHEMA) (string, error) {
	modelId := "us.anthropic.claude-3-5-haiku-20241022-v1:0"
	ctx := context.TODO()

	// past messages
	var contentBlocks []types.Message

	// add past messages
	for _, msg := range pastMessages {
		contentBlocks = append(contentBlocks, types.Message{
			Role: types.ConversationRole(msg.Role),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: fmt.Sprintf("%v", msg.Content),
				},
			},
		})
	}

	contentBlocks = append(contentBlocks, types.Message{
		Role: types.ConversationRole("user"),
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{
				Value: fmt.Sprintf("%v", request),
			},
		},
	})

	var schemaObj map[string]any
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		log.Println(err)
		return "", err
	}

	schemaDoc := document.NewLazyDocument(schemaObj)

	output, err := wrapper.BedrockRuntimeClient.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId:  aws.String(modelId),
		Messages: contentBlocks,
		System: []types.SystemContentBlock{&types.SystemContentBlockMemberText{
			Value: system,
		}},
		ToolConfig: &types.ToolConfiguration{
			Tools: []types.Tool{
				&types.ToolMemberToolSpec{
					Value: types.ToolSpecification{
						InputSchema: &types.ToolInputSchemaMemberJson{
							Value: schemaDoc,
						},
						Name:        aws.String("JSON_Output"),
						Description: aws.String("Generate structured output"),
					},
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("model err: %s : %w", modelId, err)
	}

	p := output.Output.(*types.ConverseOutputMemberMessage)

	// type switches can be used to check the union value
	union := p.Value.Content[0]

	switch v := union.(type) {
	case *types.ContentBlockMemberCachePoint:
		_ = v.Value // Value is types.CachePointBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberCitationsContent:
		_ = v.Value // Value is types.CitationsContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberDocument:
		_ = v.Value // Value is types.DocumentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberGuardContent:
		_ = v.Value // Value is types.GuardrailConverseContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil
	case *types.ContentBlockMemberImage:
		_ = v.Value // Value is types.ImageBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberReasoningContent:
		_ = v.Value // Value is types.ReasoningContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberText:
		_ = v.Value // Value is string
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberToolResult:
		_ = v.Value // Value is types.ToolResultBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberToolUse: // seems to be the one returned
		_ = v.Value // Value is types.ToolUseBlock

		bob, err := v.Value.Input.MarshalSmithyDocument()
		if err != nil {
			log.Fatalln(err)
		}
		return string(bob), nil

	case *types.ContentBlockMemberVideo:
		_ = v.Value // Value is types.VideoBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.UnknownUnionMember:
		return fmt.Sprintf("info: %+v", v.Value), nil
	default:
		return "", fmt.Errorf("problem with ai schema")

	}

}

func (wrapper BedrockClient) GenerateAISchema(request string, system string, pastMessages []map[string]string, schema string) (string, error) {
	modelId := MODEL
	ctx := context.TODO()

	// past messages
	var contentBlocks []types.Message

	// add past messages
	for _, msg := range pastMessages {
		contentBlocks = append(contentBlocks, types.Message{
			Role: types.ConversationRole(msg["role"]),
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{
					Value: fmt.Sprintf("%v", msg["content"]),
				},
			},
		})
	}

	contentBlocks = append(contentBlocks, types.Message{
		Role: types.ConversationRole("user"),
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{
				Value: fmt.Sprintf("%v", request),
			},
		},
	})

	var schemaObj map[string]any
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		log.Println(err)
		return "", err
	}

	schemaDoc := document.NewLazyDocument(schemaObj)

	output, err := wrapper.BedrockRuntimeClient.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId:  aws.String(modelId),
		Messages: contentBlocks,
		System: []types.SystemContentBlock{&types.SystemContentBlockMemberText{
			Value: system,
		}},
		ToolConfig: &types.ToolConfiguration{
			Tools: []types.Tool{
				&types.ToolMemberToolSpec{
					Value: types.ToolSpecification{
						InputSchema: &types.ToolInputSchemaMemberJson{
							Value: schemaDoc,
						},
						Name:        aws.String("API_Connector"),
						Description: aws.String("Tool to generate payload to call vendor APIs"),
					},
				},
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("model err: %s : %w", modelId, err)
	}

	p := output.Output.(*types.ConverseOutputMemberMessage)

	// type switches can be used to check the union value
	union := p.Value.Content[0]

	switch v := union.(type) {
	case *types.ContentBlockMemberCachePoint:
		_ = v.Value // Value is types.CachePointBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberCitationsContent:
		_ = v.Value // Value is types.CitationsContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberDocument:
		_ = v.Value // Value is types.DocumentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberGuardContent:
		_ = v.Value // Value is types.GuardrailConverseContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil
	case *types.ContentBlockMemberImage:
		_ = v.Value // Value is types.ImageBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberReasoningContent:
		_ = v.Value // Value is types.ReasoningContentBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberText:
		_ = v.Value // Value is string
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberToolResult:
		_ = v.Value // Value is types.ToolResultBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.ContentBlockMemberToolUse: // seems to be the one returned
		_ = v.Value // Value is types.ToolUseBlock

		bob, err := v.Value.Input.MarshalSmithyDocument()
		if err != nil {
			log.Fatalln(err)
		}
		return string(bob), nil

	case *types.ContentBlockMemberVideo:
		_ = v.Value // Value is types.VideoBlock
		return fmt.Sprintf("info: %+v", v.Value), nil

	case *types.UnknownUnionMember:
		return fmt.Sprintf("info: %+v", v.Value), nil
	default:
		return "", fmt.Errorf("problem with ai schema")

	}

}
