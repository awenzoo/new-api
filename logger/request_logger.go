package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
)

var (
	// RequestLogEnabled 是否启用请求报文记录，通过环境变量 REQUEST_LOG_ENABLED=true 开启
	RequestLogEnabled = os.Getenv("REQUEST_LOG_ENABLED") == "true"

	requestLogDir     string
	requestLogDirOnce sync.Once
)

const (
	// maxLogFilesPerUser 每个用户最多保留的日志文件数量
	maxLogFilesPerUser = 10

	// maxLogFileSize 单个日志文件最大大小 (20MB)
	maxLogFileSize = 20 * 1024 * 1024
)

// getRequestLogDir 获取请求报文日志目录
func getRequestLogDir() string {
	requestLogDirOnce.Do(func() {
		requestLogDir = filepath.Join(*common.LogDir, "request_logs")
		_ = os.MkdirAll(requestLogDir, 0755)
	})
	return requestLogDir
}

// rotateIfNeeded 检查并执行日志轮转：超过大小限制则创建新文件，超过文件数量则删除最旧的
func rotateIfNeeded(userDir string) (*os.File, error) {
	_ = os.MkdirAll(userDir, 0755)

	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, err
	}

	var logFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, filepath.Join(userDir, entry.Name()))
		}
	}

	if len(logFiles) == 0 {
		fileName := filepath.Join(userDir, fmt.Sprintf("request_%s.log", time.Now().Format("20060102_150405")))
		return os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	// 按修改时间排序（最新的在最后）
	sort.Slice(logFiles, func(i, j int) bool {
		fi, _ := os.Stat(logFiles[i])
		fj, _ := os.Stat(logFiles[j])
		if fi == nil {
			return true
		}
		if fj == nil {
			return false
		}
		return fi.ModTime().Before(fj.ModTime())
	})

	// 检查最新文件大小
	latestFile := logFiles[len(logFiles)-1]
	info, err := os.Stat(latestFile)
	if err != nil {
		return os.OpenFile(latestFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	var targetFile string
	if info.Size() >= maxLogFileSize {
		targetFile = filepath.Join(userDir, fmt.Sprintf("request_%s.log", time.Now().Format("20060102_150405")))
	} else {
		targetFile = latestFile
	}

	// 超过文件数量限制，删除最旧的
	if len(logFiles) >= maxLogFilesPerUser {
		toDelete := len(logFiles) - maxLogFilesPerUser + 1
		for i := 0; i < toDelete && i < len(logFiles); i++ {
			_ = os.Remove(logFiles[i])
		}
	}

	return os.OpenFile(targetFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// LogRequestResponse 记录用户请求和响应报文到日志文件（异步执行）
func LogRequestResponse(userId int, requestBody []byte, responseBody []byte, reqPath string, statusCode int) {
	if !RequestLogEnabled {
		return
	}

	gopool.Go(func() {
		writeRequestLog(userId, requestBody, responseBody, reqPath, statusCode)
	})
}

func writeRequestLog(userId int, requestBody []byte, responseBody []byte, reqPath string, statusCode int) {
	userDir := filepath.Join(getRequestLogDir(), fmt.Sprintf("user_%d", userId))
	f, err := rotateIfNeeded(userDir)
	if err != nil {
		common.SysError(fmt.Sprintf("failed to open request log file for user %d: %s", userId, err.Error()))
		return
	}
	defer f.Close()

	now := time.Now().Format("2006/01/02 - 15:04:05")
	separator := strings.Repeat("=", 80)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n%s\n", separator))
	sb.WriteString(fmt.Sprintf("[%s] userId=%d path=%s status=%d\n", now, userId, reqPath, statusCode))
	sb.WriteString(fmt.Sprintf("%s\n", separator))

	sb.WriteString("\n>>> REQUEST BODY >>>\n")
	if len(requestBody) > 0 {
		sb.WriteString(formatJSON(requestBody))
	} else {
		sb.WriteString("(empty)")
	}
	sb.WriteString("\n")

	sb.WriteString("\n<<< RESPONSE BODY <<<\n")
	if len(responseBody) > 0 {
		sb.WriteString(formatJSON(responseBody))
	} else {
		sb.WriteString("(empty)")
	}
	sb.WriteString("\n\n")

	_, _ = f.WriteString(sb.String())
}

func formatJSON(data []byte) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err == nil {
		return buf.String()
	}
	return string(data)
}

// ResponseCaptureWriter 包装 gin.ResponseWriter 以捕获非流式响应体。
// 注意：仅用于非流式请求，流式 SSE 响不应使用此包装器，否则会导致内存泄漏。
type ResponseCaptureWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

// NewResponseCaptureWriter 创建响应捕获包装器
func NewResponseCaptureWriter(w gin.ResponseWriter) *ResponseCaptureWriter {
	return &ResponseCaptureWriter{
		ResponseWriter: w,
		body:           bytes.NewBufferString(""),
		status:         http.StatusOK,
	}
}

// Write 捕获写入的响应数据并转发给原始 ResponseWriter
func (w *ResponseCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteHeader 记录状态码并转发
func (w *ResponseCaptureWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Hijack 支持 WebSocket 协议升级
func (w *ResponseCaptureWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
}

// Flush 支持 SSE 流式响应
func (w *ResponseCaptureWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Status 获取捕获的 HTTP 状态码
func (w *ResponseCaptureWriter) Status() int {
	return w.status
}

// GetCapturedBody 获取捕获的响应体
func (w *ResponseCaptureWriter) GetCapturedBody() []byte {
	return w.body.Bytes()
}

// IsStreamResponse 检查响应是否为流式（SSE）
func IsStreamResponse(c *gin.Context) bool {
	contentType := c.Writer.Header().Get("Content-Type")
	return strings.HasPrefix(contentType, "text/event-stream")
}
