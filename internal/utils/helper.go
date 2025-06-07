package utils

import (
	"strings"
)

// logNo 추출 함수 개선
func ExtractLogNo(href string) string {
	if strings.Contains(href, "logNo=") {
		parts := strings.Split(href, "logNo=")
		if len(parts) > 1 {
			logNo := parts[1]
			if idx := strings.Index(logNo, "&"); idx != -1 {
				logNo = logNo[:idx]
			}
			return logNo
		}
	}

	// URL 패턴에서 숫자 ID 추출 시도
	if strings.Contains(href, "/") {
		parts := strings.Split(href, "/")
		for _, part := range parts {
			if len(part) > 5 && IsNumeric(part) {
				return part
			}
		}
	}

	return ""
}

// 숫자인지 확인하는 헬퍼 함수
func IsNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// 문자열 정리 함수 개선
func CleanText(text string) string {
	if text == "" {
		return ""
	}

	// 탭, 개행 등을 공백으로 변환
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// 연속된 공백을 하나로
	text = strings.Join(strings.Fields(text), " ")

	// 앞뒤 공백 제거
	text = strings.TrimSpace(text)

	return text
}

// 문자열을 특정 길이로 자르고 "..."을 추가하는 헬퍼 함수
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
