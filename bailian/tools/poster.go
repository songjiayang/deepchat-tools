package tools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"

	"github.com/songjiayang/deepchat-tools/bailian/sdk/poster"
)

func NewPosterTool() *protocol.Tool {
	return &protocol.Tool{
		Name:        "海报生成",
		Description: "根据用户描述生成对应风格的海报",
		InputSchema: protocol.InputSchema{
			Type: protocol.Object,
			Properties: map[string]interface{}{
				"title": map[string]string{
					"type":        "string",
					"description": "标题",
				},
				"sub_title": map[string]string{
					"type":        "string",
					"description": "副标题",
				},
				"body_text": map[string]string{
					"type":        "string",
					"description": "主要内容",
				},
				"prompt_text_zh": map[string]string{
					"type":        "string",
					"description": "文生图关键提示词",
				},
				"wh_ratios": map[string]string{
					"type":        "string",
					"description": "出图画面横竖版，默认为竖版",
				},
				"lora_name": map[string]string{
					"type":        "string",
					"description": "文生图lora名称，默认为空字符串",
				},
			},
			Required: []string{"title", "sub_title", "body_text", "prompt_text_zh"},
		},
	}
}

var (
	posterClinet = poster.NewClient(os.Getenv("DASHSCOPE_API_KEY"))
)

func NewPosterToolHandler() server.ToolHandlerFunc {
	return func(request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		params := request.Arguments
		log.Printf("input params is %#v", params)

		taskInput := newPosterTaskInput(params)
		taskID, err := posterClinet.SendRequest(taskInput)
		if err != nil {
			log.Printf("create job with error: %v", err)
			return nil, err
		}
		log.Printf("send generate task with id %#v \n", taskID)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ctx.Done():
				return nil, errors.New("resolve task response timeout")
			case <-ticker.C:
				resp, err := posterClinet.GetTaskStatus(taskID)
				if err != nil {
					return nil, err
				}

				status := resp.Output.TaskStatus

				if status == "SUCCEEDED" {
					imageUrl := resp.Output.RenderURLs[0]
					log.Printf("图片生成成功，url=%s", imageUrl)
					return &protocol.CallToolResult{
						Content: []protocol.Content{
							protocol.TextContent{
								Type: "text",
								Text: fmt.Sprintf("生成的图片 url 为 \"%s\"", imageUrl),
							},
						},
					}, nil

				} else if status == "FAILED" {
					log.Printf("任务生成失败，error: %v", resp.Output.Message)
					return nil, fmt.Errorf("任务生成失败 %s", resp.Output.Message)
				}
			}
		}
	}
}

func newPosterTaskInput(params map[string]interface{}) *poster.TaskInput {
	taskInput := &poster.TaskInput{
		GenerateMode: "generate",
		LoraWeight:   0.8,
		GenerateNum:  1,
		CtrlRatio:    0.7,
		CtrlStep:     0.7,
		WhRatios:     "竖版",
	}

	if params["title"] != nil {
		taskInput.Title = params["title"].(string)
	}
	if params["sub_title"] != nil {
		taskInput.SubTitle = params["sub_title"].(string)
	}
	if params["body_text"] != nil {
		taskInput.BodyText = params["body_text"].(string)
	}
	if params["prompt_text_zh"] != nil {
		taskInput.PromptTextZh = params["prompt_text_zh"].(string)
	}
	if params["lora_name"] != nil {
		taskInput.LoraName = params["lora_name"].(string)
	}
	if params["wh_ratios"] != nil {
		taskInput.WhRatios = params["wh_ratios"].(string)
	}

	return taskInput
}
