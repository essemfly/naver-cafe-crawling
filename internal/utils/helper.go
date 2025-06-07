package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

// 여러 URL을 시도하며 HTML 문서를 가져오는 함수
func TryGetDocument(urls []string, client *http.Client) (*goquery.Document, string, error) {
	for _, url := range urls {
		doc, err := GetHTMLResponse(client, url)
		if err == nil {
			return doc, url, nil
		}
	}
	return nil, "", fmt.Errorf("모든 URL 시도 실패")
}

// 여러 셀렉터 중 첫 번째 매칭을 찾는 함수
func FindFirstMatch(doc *goquery.Document, selectors string) string {
	return CleanText(doc.Find(selectors).First().Text())
}

// 게시글 내용을 추출하는 함수
func ExtractContent(doc *goquery.Document, contentSelectors string) string {
	var contentBuilder strings.Builder
	doc.Find(contentSelectors).Each(func(i int, s *goquery.Selection) {
		s.Find("img, .se-sticker, .se-module-oglink, .se-map-container, .se-file-block, script, style").Remove()
		text := CleanText(s.Text())
		if text != "" && len(text) > 10 {
			contentBuilder.WriteString(text)
			contentBuilder.WriteString("\n")
		}
	})
	return CleanText(contentBuilder.String())
}

// 링크와 제목을 추출하는 함수
func ExtractLinkAndTitle(s *goquery.Selection, linkSelectors string) (string, string) {
	link := s.Find(linkSelectors).First()
	if link.Length() == 0 {
		return "", ""
	}
	href, _ := link.Attr("href")
	return href, strings.TrimSpace(link.Text())
}
