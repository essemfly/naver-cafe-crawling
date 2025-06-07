package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"naverCafeCrawler/internal/crawling"
	"naverCafeCrawler/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// HTTP 클라이언트 설정 (재사용)
var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
	Timeout: 30 * time.Second, // 타임아웃 증가
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}

// 블로그 크롤링 메인 함수 - 개선된 버전
func CrawlBlog(blogID string, maxPages int) ([]crawling.BlogPost, error) {
	log.Printf("🚀 네이버 블로그 '%s' 크롤링 시작...", blogID)

	var allPosts []crawling.BlogPost
	var mu sync.Mutex

	outputDir := "output_blog"
	aiReadyDir := filepath.Join(outputDir, "ai_ready")

	// 출력 디렉토리 생성
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("출력 디렉토리 생성 실패: %v", err)
	}
	if err := os.MkdirAll(aiReadyDir, 0755); err != nil {
		return nil, fmt.Errorf("AI Ready 디렉토리 생성 실패: %v", err)
	}

	// 순차 처리로 변경 (네이버 블로그는 동시 요청에 민감할 수 있음)
	for page := 1; page <= maxPages; page++ {
		log.Printf("🔄 %d/%d 페이지 처리 중...", page, maxPages)

		postsOnPage, err := crawling.GetBlogPostList(blogID, page)
		if err != nil {
			log.Printf("⚠️ 페이지 %d 게시글 목록 가져오기 실패: %v", page, err)
			continue // 실패해도 다음 페이지 진행
		}

		if len(postsOnPage) == 0 {
			log.Printf("⚠️ 페이지 %d에서 게시글을 찾을 수 없습니다.", page)
			continue
		}

		var detailedPostsOnPage []crawling.BlogPost
		for i, post := range postsOnPage {
			log.Printf("  📖 %d페이지 게시글 %d/%d 상세 정보 처리 중... (ID: %s)", page, i+1, len(postsOnPage), post.ID)

			detail, err := crawling.GetBlogPostDetail(blogID, post.ID)
			if err != nil {
				log.Printf("⚠️ 게시글 %s 상세 정보 가져오기 실패: %v", post.ID, err)
				continue
			}

			// 빈 게시글은 제외
			if detail.Title != "" || detail.Content != "" {
				detailedPostsOnPage = append(detailedPostsOnPage, detail)
			}
		}

		if len(detailedPostsOnPage) == 0 {
			log.Printf("⚠️ 페이지 %d에서 유효한 게시글을 찾을 수 없습니다.", page)
			continue
		}

		mu.Lock()
		allPosts = append(allPosts, detailedPostsOnPage...)
		mu.Unlock()

		// 페이지별 저장
		timestamp := time.Now().Format("20060102_150405")
		pageFilename := filepath.Join(outputDir, fmt.Sprintf("blog_%s_page_%d_%s.json", blogID, page, timestamp))

		if err := utils.SaveToJSON(detailedPostsOnPage, pageFilename); err != nil {
			log.Printf("⚠️ %d페이지 결과 저장 실패: %v", page, err)
		}

		// AI Ready 형식으로 저장
		var aiReadyPosts []map[string]interface{}
		for _, post := range detailedPostsOnPage {
			aiReadyPost := map[string]interface{}{
				"title":   post.Title,
				"content": post.Content,
				"metadata": map[string]interface{}{
					"id":         post.ID,
					"writer":     post.Writer,
					"write_date": post.WriteDate,
					"url":        post.OriginalURL,
				},
				"comments": post.Comments,
			}
			aiReadyPosts = append(aiReadyPosts, aiReadyPost)
		}

		aiReadyFilename := filepath.Join(aiReadyDir, fmt.Sprintf("blog_%s_page_%d_%s_ai_ready.json", blogID, page, timestamp))
		if err := utils.SaveToJSON(aiReadyPosts, aiReadyFilename); err != nil {
			log.Printf("⚠️ %d페이지 AI Ready 결과 저장 실패: %v", page, err)
		}

		log.Printf("✅ %d/%d 페이지 크롤링 완료 (수집 게시글 %d개)", page, maxPages, len(detailedPostsOnPage))
	}

	// 전체 결과 저장
	if len(allPosts) > 0 {
		timestamp := time.Now().Format("20060102_150405")
		fullFilename := filepath.Join(outputDir, fmt.Sprintf("blog_%s_full_%s.json", blogID, timestamp))

		if err := utils.SaveToJSON(allPosts, fullFilename); err != nil {
			log.Printf("⚠️ 전체 결과 저장 실패: %v", err)
		}

		// AI Ready 전체 결과 저장
		var allAiReadyPosts []map[string]interface{}
		for _, post := range allPosts {
			aiReadyPost := map[string]interface{}{
				"title":   post.Title,
				"content": post.Content,
				"metadata": map[string]interface{}{
					"id":         post.ID,
					"writer":     post.Writer,
					"write_date": post.WriteDate,
					"url":        post.OriginalURL,
				},
				"comments": post.Comments,
			}
			allAiReadyPosts = append(allAiReadyPosts, aiReadyPost)
		}

		aiReadyFullFilename := filepath.Join(aiReadyDir, fmt.Sprintf("blog_%s_full_%s_ai_ready.json", blogID, timestamp))
		if err := utils.SaveToJSON(allAiReadyPosts, aiReadyFullFilename); err != nil {
			log.Printf("⚠️ AI Ready 전체 결과 저장 실패: %v", err)
		}
	}

	log.Printf("🎉 네이버 블로그 '%s' 크롤링 완료! 총 %d개 게시글 수집", blogID, len(allPosts))
	return allPosts, nil
}

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Error loading .env file:", err)
		return
	}

	blogID := os.Getenv("NAVER_BLOG_ID")
	if blogID == "" {
		log.Fatal("NAVER_BLOG_ID 환경 변수가 설정되지 않았습니다.")
	}

	maxPages := 3

	log.Printf("🎯 대상 블로그: %s", blogID)
	log.Printf("📄 크롤링 페이지 수: %d", maxPages)

	posts, err := CrawlBlog(blogID, maxPages)
	if err != nil {
		log.Fatal("❌ 크롤링 중 오류 발생:", err)
	}

	fmt.Printf("✅ 크롤링 완료! 총 %d개 블로그 게시글 수집\n", len(posts))

	// 결과 요약 출력
	if len(posts) > 0 {
		fmt.Printf("\n📊 수집 결과 요약:\n")
		for i, post := range posts {
			if i >= 5 { // 최대 5개만 출력
				fmt.Printf("... 외 %d개 게시글\n", len(posts)-5)
				break
			}
			fmt.Printf("📌 [%d] %s\n", i+1, post.Title)
			fmt.Printf("   👤 %s | 📅 %s | 💬 %d개 댓글\n", post.Writer, post.WriteDate, len(post.Comments))
			fmt.Printf("   📝 %s...\n", utils.TruncateString(post.Content, 100))
			fmt.Println()
		}
	} else {
		fmt.Println("⚠️ 수집된 게시글이 없습니다. 블로그 ID를 확인해주세요.")
	}
}
