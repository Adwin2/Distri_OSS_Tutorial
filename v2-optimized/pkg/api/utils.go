package api

import (
	"encoding/json"
	"io"
	"net/http"
)

// ResponseWriter 包装的ResponseWriter，用于捕获状态码
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader 重写WriteHeader方法
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// WriteJSON 写入JSON响应
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// WriteError 写入错误响应
func WriteError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	errorResponse := map[string]interface{}{
		"error":   message,
		"code":    code,
		"success": false,
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}

// WriteSuccess 写入成功响应
func WriteSuccess(w http.ResponseWriter, data interface{}) {
	response := map[string]interface{}{
		"data":    data,
		"success": true,
	}
	
	WriteJSON(w, response)
}

// CopyResponse 复制响应体
func CopyResponse(w http.ResponseWriter, src io.Reader) (int64, error) {
	return io.Copy(w, src)
}

// ParseJSON 解析JSON请求体
func ParseJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
