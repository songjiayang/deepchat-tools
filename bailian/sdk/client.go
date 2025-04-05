package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	baseURL       = "https://dashscope.aliyuncs.com/api/v1"
	headerAsync   = "X-DashScope-Async"
	headerContent = "Content-Type"
	headerAuth    = "Authorization"
	contentJSON   = "application/json"
)

type Client struct {
	apiKey string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       2,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: true,
			},
		},
	}
}

type TaskRequest struct {
	Model      string                 `json:"model"`
	Input      interface{}            `json:"input"`
	Parameters map[string]interface{} `json:"parameters"`
}

type TaskResponse struct {
	Output struct {
		TaskID     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
	} `json:"output"`
	RequestId string `json:"request_id"`
	Code      string `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (resp TaskResponse) Error() error {
	if resp.Code == "" {
		return nil
	}

	return fmt.Errorf("new task with error: <code:%s, message: %s>", resp.Code, resp.Message)
}

type TaskStatusResponse struct {
	RequestId string `json:"request_id"`
	Output    struct {
		TaskID        string   `json:"task_id"`
		TaskStatus    string   `json:"task_status"`
		SubmitTime    string   `json:"submit_time"`
		ScheduledTime string   `json:"scheduled_time"`
		EndTime       string   `json:"end_time"`
		RenderURLs    []string `json:"render_urls"` // for poster
		Code          string   `json:"code,omitempty"`
		Message       string   `json:"message,omitempty"`
	} `json:"output"`
	Usage struct {
		ImageCount int `json:"image_count"`
	} `json:"usage"`
}

func (c *Client) SendRequest(path string, input *TaskRequest) (string, error) {
	body, _ := json.Marshal(input)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/services/aigc/%s", baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(headerAsync, "enable")
	req.Header.Set(headerContent, contentJSON)
	req.Header.Set(headerAuth, fmt.Sprintf("Bearer %s", c.apiKey))

	respBody, err := c.Do(req)
	if err != nil {
		return "", err
	}

	var resp TaskResponse
	if err = json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("unmarshal new task response with error; %v", err)
	}

	return resp.Output.TaskID, resp.Error()
}

func (c *Client) GetTaskStatus(taskID string) (*TaskStatusResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tasks/%s", baseURL, taskID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set(headerContent, contentJSON)
	req.Header.Set(headerAuth, fmt.Sprintf("Bearer %s", c.apiKey))

	respBody, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	var response TaskStatusResponse
	if err = json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func (c *Client) GetTaskStatusWithTimeout(taskID string, timeout time.Duration) (*TaskStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("resolve task response timeout")
		case <-ticker.C:
			resp, err := c.GetTaskStatus(taskID)
			if err != nil {
				return nil, err
			}

			status := resp.Output.TaskStatus
			switch status {
			case TaskStatusSucceeded:
				return resp, err
			case TaskStatusFailed:
				log.Printf("任务生成失败，error: %v", resp.Output.Message)
				return nil, fmt.Errorf("任务生成失败 %s", resp.Output.Message)
			default:
				log.Printf("任务当前处于 %s 状态", status)
			}
		}
	}
}

func (c *Client) Do(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, nil
}
