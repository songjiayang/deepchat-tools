package poster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL       = "https://dashscope.aliyuncs.com"
	defaultModel  = "wanx-poster-generation-v1"
	headerAsync   = "X-DashScope-Async"
	headerContent = "Content-Type"
	headerAuth    = "Authorization"
	contentJSON   = "application/json"
)

// Client represents the Aliyun client.
type Client struct {
	apiKey string
}

// NewClient creates a new Aliyun client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// Reference API https://help.aliyun.com/zh/model-studio/creative-poster-generation
// TaskInput represents the input parameters for the task generation request.
type TaskInput struct {
	Title        string  `json:"title"`
	SubTitle     string  `json:"sub_title"`
	BodyText     string  `json:"body_text"`
	PromptTextZh string  `json:"prompt_text_zh"`
	WhRatios     string  `json:"wh_ratios"`
	LoraName     string  `json:"lora_name"`
	LoraWeight   float64 `json:"lora_weight"`
	CtrlRatio    float64 `json:"ctrl_ratio"`
	CtrlStep     float64 `json:"ctrl_step"`
	GenerateMode string  `json:"generate_mode"`
	GenerateNum  int     `json:"generate_num"`
}

// TaskResponse represents the response from the task generation request.
type TaskResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Code       string `json:"code,omitempty"`
		Message    string `json:"message,omitempty"`
	} `json:"output"`
	RequestId string `json:"request_id"`
}

// TaskStatusResponse represents the response from the task status query request.
type TaskStatusResponse struct {
	RequestId string `json:"request_id"`
	Output    struct {
		TaskID              string   `json:"task_id"`
		TaskStatus          string   `json:"task_status"`
		SubmitTime          string   `json:"submit_time"`
		ScheduledTime       string   `json:"scheduled_time"`
		EndTime             string   `json:"end_time"`
		RenderURLs          []string `json:"render_urls"`
		AuxiliaryParameters []string `json:"auxiliary_parameters"`
		BGURLs              []string `json:"bg_urls"`
		Code                string   `json:"code,omitempty"`
		Message             string   `json:"message,omitempty"`
	} `json:"output"`
	Usage struct {
		ImageCount int `json:"image_count"`
	} `json:"usage"`
}

// SendRequest sends a request to the Aliyun service to generate a task.
func (c *Client) SendRequest(input *TaskInput) (string, error) {
	body, err := json.Marshal(map[string]interface{}{
		"model":      defaultModel,
		"input":      input,
		"parameters": map[string]interface{}{},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/services/aigc/text2image/image-synthesis", baseURL), bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(headerAsync, "enable")
	req.Header.Set(headerContent, contentJSON)
	req.Header.Set(headerAuth, fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			code, _ := errorResp["code"].(string)
			message, _ := errorResp["message"].(string)
			requestID, _ := errorResp["request_id"].(string)
			return "", fmt.Errorf("request failed with status code: %d, code: %s, message: %s, request_id: %s", resp.StatusCode, code, message, requestID)
		}
		return "", fmt.Errorf("request failed with status code: %d, response: %s", resp.StatusCode, respBody)
	}

	var response TaskResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Output.TaskStatus != "PENDING" {
		if response.Output.TaskStatus == "FAILED" {
			return "", fmt.Errorf("task generation failed: code: %s, message: %s", response.Output.Code, response.Output.Message)
		}
		return "", fmt.Errorf("unexpected task status: %s", response.Output.TaskStatus)
	}

	return response.Output.TaskID, nil
}

// GetTaskStatus sends a request to the Aliyun service to get the status of a task.
func (c *Client) GetTaskStatus(taskID string) (*TaskStatusResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/tasks/%s", baseURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(headerContent, contentJSON)
	req.Header.Set(headerAuth, fmt.Sprintf("Bearer %s", c.apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			code, _ := errorResp["code"].(string)
			message, _ := errorResp["message"].(string)
			requestID, _ := errorResp["request_id"].(string)
			return nil, fmt.Errorf("request failed with status code: %d, code: %s, message: %s, request_id: %s", resp.StatusCode, code, message, requestID)
		}
		return nil, fmt.Errorf("request failed with status code: %d, response: %s", resp.StatusCode, respBody)
	}

	var response TaskStatusResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
