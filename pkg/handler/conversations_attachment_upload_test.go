package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newAttachmentUploadTestRequest(channelID, filePath string, extra map[string]any) mcp.CallToolRequest {
	arguments := map[string]any{
		"channel_id": channelID,
		"file_path":  filePath,
	}
	for key, value := range extra {
		arguments[key] = value
	}

	request := mcp.CallToolRequest{}
	request.Params.Name = "attachment_upload"
	request.Params.Arguments = arguments
	return request
}

func createUploadTempFile(t *testing.T, size int64) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "attachment-upload-*.txt")
	require.NoError(t, err)

	if size > 0 {
		require.NoError(t, file.Truncate(size))
	}
	require.NoError(t, file.Close())

	return file.Name()
}

func newAttachmentUploadTestHandler() *ConversationsHandler {
	return &ConversationsHandler{
		logger: zap.NewNop(),
	}
}

func TestUnitParseParamsToolFilesUpload_DisabledByDefault(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)
	request := newAttachmentUploadTestRequest("D123456", filePath, nil)

	_, err := handler.parseParamsToolFilesUpload(context.Background(), request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_upload tool is disabled")
}

func TestUnitParseParamsToolFilesUpload_EnabledViaEnabledTools(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "conversations_history,attachment_upload")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)
	request := newAttachmentUploadTestRequest("D123456", filePath, nil)

	params, err := handler.parseParamsToolFilesUpload(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, "D123456", params.channel)
	assert.Equal(t, filePath, params.filePath)
	assert.Equal(t, filepath.Base(filePath), params.filename)
}

func TestUnitParseParamsToolFilesUpload_AllowlistPolicy(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "D123456,D654321")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)

	allowedReq := newAttachmentUploadTestRequest("D123456", filePath, nil)
	allowedParams, err := handler.parseParamsToolFilesUpload(context.Background(), allowedReq)
	require.NoError(t, err)
	assert.Equal(t, "D123456", allowedParams.channel)

	blockedReq := newAttachmentUploadTestRequest("D999999", filePath, nil)
	_, err = handler.parseParamsToolFilesUpload(context.Background(), blockedReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_upload tool is not allowed for channel")
}

func TestUnitParseParamsToolFilesUpload_BlocklistPolicy(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "!D123456")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)

	blockedReq := newAttachmentUploadTestRequest("D123456", filePath, nil)
	_, err := handler.parseParamsToolFilesUpload(context.Background(), blockedReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "attachment_upload tool is not allowed for channel")

	allowedReq := newAttachmentUploadTestRequest("D777777", filePath, nil)
	allowedParams, err := handler.parseParamsToolFilesUpload(context.Background(), allowedReq)
	require.NoError(t, err)
	assert.Equal(t, "D777777", allowedParams.channel)
}

func TestUnitParseParamsToolFilesUpload_ThreadTimestampValidation(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "true")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)
	request := newAttachmentUploadTestRequest("D123456", filePath, map[string]any{
		"thread_ts": "1234567890",
	})

	_, err := handler.parseParamsToolFilesUpload(context.Background(), request)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "thread_ts must be a valid timestamp")
}

func TestUnitParseParamsToolFilesUpload_FileValidation(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "true")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()

	directoryReq := newAttachmentUploadTestRequest("D123456", t.TempDir(), nil)
	_, err := handler.parseParamsToolFilesUpload(context.Background(), directoryReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")

	emptyFilePath := createUploadTempFile(t, 0)
	emptyReq := newAttachmentUploadTestRequest("D123456", emptyFilePath, nil)
	_, err = handler.parseParamsToolFilesUpload(context.Background(), emptyReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is empty")

	tooLargeFilePath := createUploadTempFile(t, maxUploadFileSizeBytes+1)
	tooLargeReq := newAttachmentUploadTestRequest("D123456", tooLargeFilePath, nil)
	_, err = handler.parseParamsToolFilesUpload(context.Background(), tooLargeReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestUnitParseParamsToolFilesUpload_UsesOptionalMetadata(t *testing.T) {
	t.Setenv("SLACK_MCP_ATTACHMENT_UPLOAD_TOOL", "true")
	t.Setenv("SLACK_MCP_ENABLED_TOOLS", "")

	handler := newAttachmentUploadTestHandler()
	filePath := createUploadTempFile(t, 32)
	request := newAttachmentUploadTestRequest("D123456", filePath, map[string]any{
		"thread_ts":       "1771753347.553909",
		"filename":        "evidence.txt",
		"title":           "Upload Test Artifact",
		"initial_comment": "Automated validation message",
	})

	params, err := handler.parseParamsToolFilesUpload(context.Background(), request)
	require.NoError(t, err)
	assert.Equal(t, "D123456", params.channel)
	assert.Equal(t, filePath, params.filePath)
	assert.Equal(t, "1771753347.553909", params.threadTs)
	assert.Equal(t, "evidence.txt", params.filename)
	assert.Equal(t, "Upload Test Artifact", params.title)
	assert.Equal(t, "Automated validation message", params.initialComment)
}
