package tools

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/songjiayang/deepchat-tools/bailian/sdk"
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
					"description": "主要内容，不超过50个字符",
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
	client = sdk.NewClient(os.Getenv("DASHSCOPE_API_KEY"))
)

func NewPosterToolHandler() server.ToolHandlerFunc {
	return func(request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		params := request.Arguments
		log.Printf("输入参数为 %#v", params)

		taskInput := newPosterTaskInput(params)
		taskID, err := client.SendPosterRequest(taskInput)
		if err != nil {
			log.Printf("create job with error: %v", err)
			return nil, err
		}
		log.Printf("任务创建成功，taskId=%s", taskID)

		resp, err := client.GetTaskStatusWithTimeout(taskID, 3*time.Minute)
		if err != nil {
			return nil, err
		}

		// response image with url
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
	}
}

func newPosterTaskInput(params map[string]interface{}) map[string]interface{} {
	// set default value
	params["generate_mode"] = "generate"
	params["lora_weight"] = 0.8
	params["ctrl_ratio"] = 0.7
	params["ctrl_step"] = 0.7
	params["generate_num"] = 1
	if params["wh_ratios"] == nil {
		params["wh_ratios"] = "竖版"
	}

	return params
}
