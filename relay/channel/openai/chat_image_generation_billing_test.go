package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newChatImageBillingTestContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	return c
}

func newChatImageBillingRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		RelayFormat: types.RelayFormatOpenAI,
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gpt-5.5-native",
		},
	}
}

func TestOpenaiHandlerMarksMessageImagesForImageGenerationBilling(t *testing.T) {
	c := newChatImageBillingTestContext()
	responseBody := `{"id":"chatcmpl_test","object":"chat.completion","created":1710000000,"model":"gpt-5.5-native","choices":[{"index":0,"message":{"role":"assistant","content":"done","images":[{"type":"image_url","quality":"low","size":"1024x1024","image_url":{"url":"data:image/png;base64,aGVsbG8="}}]},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}

	usage, apiErr := OpenaiHandler(c, newChatImageBillingRelayInfo(), resp)

	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	require.True(t, c.GetBool("image_generation_call"))
	require.Equal(t, "low", c.GetString("image_generation_call_quality"))
	require.Equal(t, "1024x1024", c.GetString("image_generation_call_size"))
}

func TestMarkChatImageGenerationBillingContextIgnoresPlainText(t *testing.T) {
	c := newChatImageBillingTestContext()
	responseBody := []byte(`{"choices":[{"message":{"role":"assistant","content":"plain text only"}}]}`)

	markChatImageGenerationBillingContext(c, responseBody)

	require.False(t, c.GetBool("image_generation_call"))
}

func TestMarkChatImageGenerationBillingContextMarksImagesWithoutMetadata(t *testing.T) {
	c := newChatImageBillingTestContext()
	responseBody := []byte(`{"choices":[{"message":{"role":"assistant","content":"done","images":[{"type":"image_url","image_url":{"url":"data:image/png;base64,aGVsbG8="}}]}}]}`)

	markChatImageGenerationBillingContext(c, responseBody)

	require.True(t, c.GetBool("image_generation_call"))
	require.Empty(t, c.GetString("image_generation_call_quality"))
	require.Empty(t, c.GetString("image_generation_call_size"))
}

func TestMarkChatStreamImageGenerationBillingContextMarksDeltaImages(t *testing.T) {
	c := newChatImageBillingTestContext()
	streamItems := []string{
		`{"id":"chatcmpl_test","object":"chat.completion.chunk","created":1710000000,"model":"gpt-5.5-native","choices":[{"index":0,"delta":{"role":"assistant","images":[{"type":"image_url","image_url":{"url":"data:image/png;base64,aGVsbG8=","quality":"high","size":"1536x1024"}}]},"finish_reason":null}]}`,
		`{"id":"chatcmpl_test","object":"chat.completion.chunk","created":1710000000,"model":"gpt-5.5-native","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}`,
	}

	markChatStreamImageGenerationBillingContext(c, streamItems)

	require.True(t, c.GetBool("image_generation_call"))
	require.Equal(t, "high", c.GetString("image_generation_call_quality"))
	require.Equal(t, "1536x1024", c.GetString("image_generation_call_size"))
}
