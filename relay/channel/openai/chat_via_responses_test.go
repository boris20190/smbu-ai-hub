package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func chatImageGenerationResponseJSON() string {
	return `{"id":"resp_test","created_at":1710000000,"model":"gpt-5.5-native","output":[{"type":"image_generation_call","quality":"low","size":"1024x1024"},{"type":"message","role":"assistant","content":[{"type":"output_text","text":"done"}]}],"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}`
}

func newChatViaResponsesTestContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	return c
}

func newChatViaResponsesRelayInfo() *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		RelayFormat: types.RelayFormatOpenAI,
		ChannelMeta: &relaycommon.ChannelMeta{
			UpstreamModelName: "gpt-5.5-native",
		},
	}
}

func TestOaiResponsesToChatHandlerMarksImageGenerationBilling(t *testing.T) {
	c := newChatViaResponsesTestContext()
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(chatImageGenerationResponseJSON())),
	}

	usage, apiErr := OaiResponsesToChatHandler(c, newChatViaResponsesRelayInfo(), resp)

	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	require.True(t, c.GetBool("image_generation_call"))
	require.Equal(t, "low", c.GetString("image_generation_call_quality"))
	require.Equal(t, "1024x1024", c.GetString("image_generation_call_size"))
}

func TestOaiResponsesToChatStreamHandlerMarksImageGenerationBilling(t *testing.T) {
	oldTimeout := constant.StreamingTimeout
	constant.StreamingTimeout = 30
	t.Cleanup(func() {
		constant.StreamingTimeout = oldTimeout
	})

	c := newChatViaResponsesTestContext()
	streamBody := "data: {\"type\":\"response.completed\",\"response\":" + chatImageGenerationResponseJSON() + "}\n\n" +
		"data: [DONE]\n\n"
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(streamBody)),
	}

	usage, apiErr := OaiResponsesToChatStreamHandler(c, newChatViaResponsesRelayInfo(), resp)

	require.Nil(t, apiErr)
	require.NotNil(t, usage)
	require.True(t, c.GetBool("image_generation_call"))
	require.Equal(t, "low", c.GetString("image_generation_call_quality"))
	require.Equal(t, "1024x1024", c.GetString("image_generation_call_size"))
}
