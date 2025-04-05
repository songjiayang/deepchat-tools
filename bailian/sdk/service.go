package sdk

const (
	posterModel = "wanx-poster-generation-v1"
	posterPath  = "text2image/image-synthesis"
)

func (c *Client) SendPosterRequest(input map[string]interface{}) (taskID string, err error) {
	return c.SendRequest(posterPath, &TaskRequest{
		Model: posterModel,
		Input: input,
	})
}
