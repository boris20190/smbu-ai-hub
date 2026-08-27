package openai

import (
	"github.com/QuantumNous/new-api/dto"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func markImageGenerationBillingContext(c *gin.Context, response *dto.OpenAIResponsesResponse) {
	if c == nil || response == nil || !response.HasImageGenerationCall() {
		return
	}

	setImageGenerationBillingContext(c, response.GetQuality(), response.GetSize())
}

func markChatImageGenerationBillingContext(c *gin.Context, responseBody []byte) {
	if c == nil || len(responseBody) == 0 {
		return
	}

	quality, size, ok := chatImageGenerationBillingMetadata(gjson.ParseBytes(responseBody))
	if !ok {
		return
	}

	setImageGenerationBillingContext(c, quality, size)
}

func markChatStreamImageGenerationBillingContext(c *gin.Context, streamItems []string) {
	if c == nil || len(streamItems) == 0 {
		return
	}

	for _, item := range streamItems {
		if item == "" {
			continue
		}
		quality, size, ok := chatImageGenerationBillingMetadata(gjson.Parse(item))
		if !ok {
			continue
		}
		setImageGenerationBillingContext(c, quality, size)
		return
	}
}

func setImageGenerationBillingContext(c *gin.Context, quality string, size string) {
	c.Set("image_generation_call", true)
	c.Set("image_generation_call_quality", quality)
	c.Set("image_generation_call_size", size)
}

func chatImageGenerationBillingMetadata(root gjson.Result) (string, string, bool) {
	choices := root.Get("choices")
	if !choices.IsArray() {
		return "", "", false
	}

	for _, choice := range choices.Array() {
		if quality, size, ok := imageGenerationMetadataFromImages(choice.Get("message.images")); ok {
			return quality, size, true
		}
		if quality, size, ok := imageGenerationMetadataFromImages(choice.Get("delta.images")); ok {
			return quality, size, true
		}
		if quality, size, ok := imageGenerationMetadataFromContentParts(choice.Get("message.content")); ok {
			return quality, size, true
		}
	}

	return "", "", false
}

func imageGenerationMetadataFromImages(images gjson.Result) (string, string, bool) {
	if !images.IsArray() {
		return "", "", false
	}

	imageItems := images.Array()
	if len(imageItems) == 0 {
		return "", "", false
	}

	for _, image := range imageItems {
		quality := firstNonEmptyString(
			image.Get("quality").String(),
			image.Get("image_url.quality").String(),
		)
		size := firstNonEmptyString(
			image.Get("size").String(),
			image.Get("image_url.size").String(),
		)
		if quality != "" || size != "" {
			return quality, size, true
		}
	}

	return "", "", true
}

func imageGenerationMetadataFromContentParts(content gjson.Result) (string, string, bool) {
	if !content.IsArray() {
		return "", "", false
	}

	for _, part := range content.Array() {
		partType := part.Get("type").String()
		if partType != "image_url" && partType != "output_image" {
			continue
		}
		quality := firstNonEmptyString(
			part.Get("quality").String(),
			part.Get("image_url.quality").String(),
		)
		size := firstNonEmptyString(
			part.Get("size").String(),
			part.Get("image_url.size").String(),
		)
		return quality, size, true
	}

	return "", "", false
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
