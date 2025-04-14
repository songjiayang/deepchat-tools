package tools

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	"github.com/tidwall/gjson"
	"github.com/volcengine/volc-sdk-golang/service/visual"
)

func init() {
	visual.DefaultInstance.Client.SetAccessKey(os.Getenv("VolcAccessKeyID"))
	visual.DefaultInstance.Client.SetSecretKey(os.Getenv("VolcSecretAccessKey"))
	visual.DefaultInstance.SetRegion("cn-north-1")
}

func NewImageStyleTool() *protocol.Tool {
	return &protocol.Tool{
		Name:        "图片风格化",
		Description: "根据输入图片文件路径进行风格化",
		InputSchema: protocol.InputSchema{
			Type: protocol.Object,
			Properties: map[string]interface{}{
				"image_path": map[string]string{
					"type":        "string",
					"description": "输入图片文件路径",
				},
				"image_style": map[string]string{
					"type":        "string",
					"description": "风格名称， 默认为 \"网红日漫风\"",
				},
			},
			Required: []string{"image_path", "image_style"},
		},
	}
}

func NewImageStyleHandler() server.ToolHandlerFunc {
	return func(request *protocol.CallToolRequest) (*protocol.CallToolResult, error) {
		params := request.Arguments
		log.Printf("输入参数为 %#v", params)

		inputImage := params["image_path"].(string)
		data, err := os.ReadFile(inputImage)
		if err != nil {
			log.Printf("read image with error: %v", err)
			return nil, err
		}

		reqKey := resolveStyleKey(params["image_style"].(string))
		inputB64Image := base64.StdEncoding.EncodeToString(data)
		reqBody := map[string]interface{}{
			"req_key":            reqKey,
			"binary_data_base64": []string{inputB64Image},
			"return_url":         true,
		}

		resp, _, err := visual.DefaultInstance.CVProcess(reqBody)
		if err != nil {
			return nil, err
		}

		// resolve image url
		jsonData, _ := json.Marshal(resp)
		imageUrl := gjson.Get(string(jsonData), "data.image_urls.0")
		// response image with url
		log.Printf("图片生成成功，url=%s", imageUrl)
		return &protocol.CallToolResult{
			Content: []protocol.Content{
				protocol.TextContent{
					Type: "text",
					Text: fmt.Sprintf("生成的图片 Url 为 \"%s\"", imageUrl),
				},
			},
		}, nil
	}
}

var allStyles = map[string]string{
	"网红日漫":  "img2img_ghibli_style",
	"3D":    "img2img_disney_3d_style",
	"写实":    "img2img_real_mix_style",
	"天使":    "img2img_pastel_boys_style",
	"动漫":    "img2img_cartoon_style",
	"日漫":    "img2img_makoto_style",
	"公主":    "img2img_rev_animated_style",
	"梦幻":    "img2img_blueline_style",
	"水墨风":   "img2img_water_ink_style",
	"新莫奈花园": " i2i_ai_create_monet",
	"水彩":    "img2img_water_paint_style",
	"莫奈花园":  "img2img_comic_style",
	"精致美漫":  "img2img_comic_style",
	"赛博机械":  "img2img_comic_style",
	"精致韩漫":  "img2img_exquisite_style",
	"国风-水墨": "img2img_pretty_style",
	"浪漫光影":  "img2img_pretty_style",
	"陶瓷娃娃":  "img2img_ceramics_style",
	"中国红":   "img2img_chinese_style",
	"丑萌粘土":  "img2img_clay_style",
	"可爱玩偶":  "img2img_clay_style",
	"动画电影":  "img2img_3d_style",
	"玩偶":    "img2img_3d_style",
}

func resolveStyleKey(imageStyle string) string {
	trimStyle := strings.TrimSuffix(imageStyle, "风格")
	trimStyle = strings.TrimSuffix(trimStyle, "风")

	if styleKey, ok := allStyles[trimStyle]; ok {
		return styleKey
	}

	log.Printf("没有找到与 \"%s\" 匹配的风格, 使用\"网红日漫风\"", imageStyle)
	return "img2img_ghibli_style"
}
