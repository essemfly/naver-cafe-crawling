package crawling

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// 요청 간 랜덤 지연을 위한 함수
func randomSleep() {
	sleepTime := time.Duration(rand.Intn(2000)+1000) * time.Millisecond
	time.Sleep(sleepTime)
}

// HTTP 클라이언트 설정
var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
	Timeout: 10 * time.Second,
}

// 응답 구조체
type ArticleListResponse struct {
	Result struct {
		ArticleList []struct {
			Type string `json:"type"`
			Item struct {
				ArticleId          int    `json:"articleId"`
				CafeId             int    `json:"cafeId"`
				Subject            string `json:"subject"`
				WriteDateTimestamp int64  `json:"writeDateTimestamp"`
				CommentCount       int    `json:"commentCount"`
				ReadCount          int    `json:"readCount"`
				LikeCount          int    `json:"likeCount"`
				WriterInfo         struct {
					NickName        string `json:"nickName"`
					MemberLevel     int    `json:"memberLevel"`
					MemberLevelName string `json:"memberLevelName"`
					Staff           bool   `json:"staff"`
					Manager         bool   `json:"manager"`
				} `json:"writerInfo"`
			} `json:"item"`
		} `json:"articleList"`
		PageInfo struct {
			LastNavigationPageNumber int  `json:"lastNavigationPageNumber"`
			VisibleNextButton        bool `json:"visibleNextButton"`
		} `json:"pageInfo"`
	} `json:"result"`
}

// 게시글 상세 응답 구조체
type ArticleDetailResponse struct {
	Result struct {
		Article struct {
			ID           int    `json:"id"`
			RefArticleID int    `json:"refArticleId"`
			ContentHtml  string `json:"contentHtml"`
			Subject      string `json:"subject"`
			WriteDate    int64  `json:"writeDate"`
			Writer       struct {
				NickName        string `json:"nickName"`
				MemberLevel     int    `json:"memberLevel"`
				MemberLevelName string `json:"memberLevelName"`
				Staff           bool   `json:"staff"`
				Manager         bool   `json:"manager"`
			} `json:"writer"`
			CommentCount int `json:"commentCount"`
			ReadCount    int `json:"readCount"`
			LikeCount    int `json:"likeCount"`
		} `json:"article"`
		Comments struct {
			Items []struct {
				ID        int    `json:"id"`
				Content   string `json:"content"`
				WriteDate int64  `json:"writeDate"`
				Writer    struct {
					NickName        string `json:"nickName"`
					MemberLevel     int    `json:"memberLevel"`
					MemberLevelName string `json:"memberLevelName"`
					Staff           bool   `json:"staff"`
					Manager         bool   `json:"manager"`
				} `json:"writer"`
				LikeCount int `json:"likeCount"`
			} `json:"items"`
		} `json:"comments"`
	} `json:"result"`
}

// HTTP 요청 보내고 응답 반환하는 함수
func getAPIResponse(url, cookie string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 필수 헤더만 설정
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", "https://cafe.naver.com")
	req.Header.Set("Origin", "https://cafe.naver.com")
	req.Header.Set("X-Cafe-Product", "pc")

	randomSleep()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP 오류: %d", resp.StatusCode)
	}

	return resp, nil
}

// 게시글 목록 가져오기
func getPostList(cafeId, boardID string, page int, pageSize int, cookie string) ([]map[string]interface{}, int, error) {
	url := fmt.Sprintf("https://apis.naver.com/cafe-web/cafe-boardlist-api/v1/cafes/%s/menus/%s/articles?page=%d&pageSize=%d&sortBy=TIME&viewType=L",
		cafeId, boardID, page, pageSize)

	resp, err := getAPIResponse(url, cookie)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var result ArticleListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	var posts []map[string]interface{}
	for _, article := range result.Result.ArticleList {
		if article.Type == "ARTICLE" {
			posts = append(posts, map[string]interface{}{
				"id":            article.Item.ArticleId,
				"title":         article.Item.Subject,
				"writer":        article.Item.WriterInfo.NickName,
				"writer_level":  article.Item.WriterInfo.MemberLevelName,
				"is_staff":      article.Item.WriterInfo.Staff,
				"is_manager":    article.Item.WriterInfo.Manager,
				"write_date":    time.Unix(article.Item.WriteDateTimestamp/1000, 0).Format("2006-01-02 15:04:05"),
				"comment_count": article.Item.CommentCount,
				"read_count":    article.Item.ReadCount,
				"like_count":    article.Item.LikeCount,
			})
		}
	}
	return posts, result.Result.PageInfo.LastNavigationPageNumber, nil
}

// 게시글 상세 정보 가져오기
func getArticleDetail(cafeId string, articleId int, cookie string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://apis.naver.com/cafe-web/cafe-articleapi/v3/cafes/%s/articles/%d?query=&useCafeId=true&requestFrom=A",
		cafeId, articleId)

	resp, err := getAPIResponse(url, cookie)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ArticleDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 게시글 정보 구성
	article := result.Result.Article
	articleDetail := map[string]interface{}{
		"id":            article.ID,
		"title":         article.Subject,
		"content_html":  article.ContentHtml,
		"writer":        article.Writer.NickName,
		"writer_level":  article.Writer.MemberLevelName,
		"is_staff":      article.Writer.Staff,
		"is_manager":    article.Writer.Manager,
		"write_date":    time.Unix(article.WriteDate/1000, 0).Format("2006-01-02 15:04:05"),
		"comment_count": article.CommentCount,
		"read_count":    article.ReadCount,
		"like_count":    article.LikeCount,
	}

	// 댓글 정보 구성
	var comments []map[string]interface{}
	for _, comment := range result.Result.Comments.Items {
		comments = append(comments, map[string]interface{}{
			"id":           comment.ID,
			"content":      comment.Content,
			"writer":       comment.Writer.NickName,
			"writer_level": comment.Writer.MemberLevelName,
			"is_staff":     comment.Writer.Staff,
			"is_manager":   comment.Writer.Manager,
			"write_date":   time.Unix(comment.WriteDate/1000, 0).Format("2006-01-02 15:04:05"),
			"like_count":   comment.LikeCount,
		})
	}
	articleDetail["comments"] = comments

	return articleDetail, nil
}

// 게시판 크롤링
func CrawlBoard(cafeId, boardID string, cookie string, maxPages int, pageSize int) ([]map[string]interface{}, error) {
	// 첫 페이지를 가져와서 마지막 페이지 번호 확인
	firstPagePosts, lastPage, err := getPostList(cafeId, boardID, 1, pageSize, cookie)
	if err != nil {
		return nil, fmt.Errorf("첫 페이지 로드 실패: %v", err)
	}

	// 크롤링할 페이지 수 결정
	pagesToCrawl := lastPage
	if maxPages > 0 && maxPages < lastPage {
		pagesToCrawl = maxPages
	}

	log.Printf("총 %d 페이지 중 %d 페이지 크롤링 시작 (페이지당 최신순, 동시 처리 3페이지)", lastPage, pagesToCrawl)

	var allPosts []map[string]interface{}
	var mu sync.Mutex

	// 첫 페이지 결과에 상세 정보 추가
	for i, post := range firstPagePosts {
		articleId := post["id"].(int)
		detail, err := getArticleDetail(cafeId, articleId, cookie)
		if err != nil {
			log.Printf("⚠️ 게시글 %d 상세 정보 가져오기 실패: %v", articleId, err)
			continue
		}
		firstPagePosts[i]["content"] = detail["content_html"]
		firstPagePosts[i]["comments"] = detail["comments"]
	}
	allPosts = append(allPosts, firstPagePosts...)

	// 컨텍스트와 에러그룹 생성
	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(3) // 동시 처리 제한

	// 2페이지부터 지정된 페이지까지 크롤링
	for page := 2; page <= pagesToCrawl; page++ {
		page := page
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				posts, _, err := getPostList(cafeId, boardID, page, pageSize, cookie)
				if err != nil {
					return fmt.Errorf("페이지 %d 크롤링 실패: %v", page, err)
				}

				// 각 게시글의 상세 정보 가져오기
				for i, post := range posts {
					articleId := post["id"].(int)
					detail, err := getArticleDetail(cafeId, articleId, cookie)
					if err != nil {
						log.Printf("⚠️ 게시글 %d 상세 정보 가져오기 실패: %v", articleId, err)
						continue
					}
					posts[i]["content"] = detail["content_html"]
					posts[i]["comments"] = detail["comments"]
				}

				mu.Lock()
				allPosts = append(allPosts, posts...)
				mu.Unlock()

				log.Printf("✅ %d/%d 페이지 크롤링 완료 (%d개 게시글)", page, pagesToCrawl, len(posts))
				return nil
			}
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	log.Printf("✅ 크롤링 완료! 총 %d개 게시글 수집", len(allPosts))
	return allPosts, nil
}
