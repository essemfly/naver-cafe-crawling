package crawling

import (
	"fmt"
	"log"
	"naverCafeCrawler/internal/utils"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Article represents a blog post.
type BlogPost struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Content     string        `json:"content"`
	Writer      string        `json:"writer"`
	WriteDate   string        `json:"write_date"`
	Comments    []BlogComment `json:"comments"`
	OriginalURL string        `json:"original_url"`
}

// BlogComment represents a comment on a blog post.
type BlogComment struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Writer    string `json:"writer"`
	WriteDate string `json:"write_date"`
}

// 게시글 목록 가져오기 - 개선된 버전
func GetBlogPostList(blogID string, page int) ([]BlogPost, error) {
	// 다양한 네이버 블로그 URL 패턴 시도
	urls := []string{
		fmt.Sprintf("https://blog.naver.com/PostList.naver?blogId=%s&from=postList&categoryNo=0&currentPage=%d", blogID, page),
		fmt.Sprintf("https://blog.naver.com/%s/postList?currentPage=%d", blogID, page),
		fmt.Sprintf("https://blog.naver.com/%s", blogID), // 메인 페이지
	}

	var doc *goquery.Document
	var err error
	var successURL string

	for _, url := range urls {
		log.Printf("🔍 시도 중인 URL: %s", url)
		doc, err = utils.GetHTMLResponse(client, url)
		if err == nil {
			successURL = url
			break
		}
		log.Printf("⚠️ URL 실패: %s - %v", url, err)
	}

	if doc == nil {
		return nil, fmt.Errorf("모든 URL 시도 실패: %v", err)
	}

	log.Printf("✅ 성공한 URL: %s", successURL)

	var posts []BlogPost

	// 다양한 셀렉터 패턴 시도
	selectors := []string{
		".post-item",
		".blog2_series",
		".item_post",
		".post_area",
		"#content-area .post",
		".list_post .post",
		".area_list_post .post",
		".blog_list .post",
	}

	found := false
	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			found = true

			// 링크 찾기
			var href string
			var title string

			// 다양한 링크 셀렉터 시도
			linkSelectors := []string{
				"a[href*='logNo=']",
				".link_post",
				".post_title a",
				".title a",
				"a.link_title",
			}

			for _, linkSel := range linkSelectors {
				if link := s.Find(linkSel).First(); link.Length() > 0 {
					if h, exists := link.Attr("href"); exists {
						href = h
						title = strings.TrimSpace(link.Text())
						break
					}
				}
			}

			// href가 상대 경로인 경우 절대 경로로 변환
			if href != "" && !strings.HasPrefix(href, "http") {
				if strings.HasPrefix(href, "/") {
					href = "https://blog.naver.com" + href
				}
			}

			// 제목이 없으면 다른 셀렉터로 찾기
			if title == "" {
				titleSelectors := []string{
					".post_title",
					".title",
					".subject",
					".tit",
				}
				for _, titleSel := range titleSelectors {
					if titleEl := s.Find(titleSel).First(); titleEl.Length() > 0 {
						title = strings.TrimSpace(titleEl.Text())
						if title != "" {
							break
						}
					}
				}
			}

			// logNo 추출
			articleID := utils.ExtractLogNo(href)

			if articleID != "" && title != "" {
				post := BlogPost{
					ID:          articleID,
					Title:       title,
					OriginalURL: fmt.Sprintf("https://blog.naver.com/%s/%s", blogID, articleID),
				}
				posts = append(posts, post)
				log.Printf("📄 발견된 게시글: ID=%s, Title=%s", articleID, title)
			}
		})

		if found {
			break
		}
	}

	// 추가적으로 iframe 내부 확인 (네이버 블로그는 iframe을 사용하는 경우가 많음)
	if !found {
		log.Printf("📝 iframe 내부 확인 중...")
		doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
			if src, exists := s.Attr("src"); exists {
				log.Printf("🔍 발견된 iframe: %s", src)
			}
		})
	}

	log.Printf("📄 페이지 %d: %d개 게시글 발견", page, len(posts))
	return posts, nil
}

// 게시글 상세 정보 가져오기 - 개선된 버전
func GetBlogPostDetail(blogID string, articleID string) (BlogPost, error) {
	// 다양한 URL 패턴 시도
	urls := []string{
		fmt.Sprintf("https://blog.naver.com/%s/%s", blogID, articleID),
		fmt.Sprintf("https://blog.naver.com/PostView.naver?blogId=%s&logNo=%s", blogID, articleID),
		fmt.Sprintf("https://blog.naver.com/PostView.nhn?blogId=%s&logNo=%s", blogID, articleID),
	}

	var doc *goquery.Document
	var err error
	var successURL string

	for _, url := range urls {
		log.Printf("🔍 게시글 URL 시도: %s", url)
		doc, err = utils.GetHTMLResponse(client, url)
		if err == nil {
			successURL = url
			break
		}
		log.Printf("⚠️ 게시글 URL 실패: %s - %v", url, err)
	}

	if doc == nil {
		return BlogPost{}, fmt.Errorf("게시글 로드 실패 (모든 URL 시도 실패): %v", err)
	}

	log.Printf("✅ 게시글 성공 URL: %s", successURL)

	var blogPost BlogPost
	blogPost.ID = articleID
	blogPost.OriginalURL = successURL

	// 제목 추출 - 다양한 셀렉터 시도
	titleSelectors := []string{
		".se-title-text",
		".pcol1 .itemSubjectBoldfont",
		".tit_area .tit",
		".post_title",
		".title_area .title",
		"#content-area .post_title",
		".se-module-text h1",
		".se-module-text h2",
		".se-module-text .se-text-paragraph:first-child",
	}

	for _, sel := range titleSelectors {
		if title := utils.CleanText(doc.Find(sel).First().Text()); title != "" {
			blogPost.Title = title
			break
		}
	}

	// 작성자 추출
	writerSelectors := []string{
		".nick_name",
		".blog_author .author_name",
		".author",
		".writer",
		".nickname",
		".blog_name",
	}

	for _, sel := range writerSelectors {
		if writer := utils.CleanText(doc.Find(sel).First().Text()); writer != "" {
			blogPost.Writer = writer
			break
		}
	}

	// 작성일 추출
	dateSelectors := []string{
		".se_time",
		".blog_header_info .date",
		"._postContents .post_info .date",
		".post_date",
		".date",
		".write_date",
	}

	for _, sel := range dateSelectors {
		if date := utils.CleanText(doc.Find(sel).First().Text()); date != "" {
			blogPost.WriteDate = date
			break
		}
	}

	// 내용 추출 - 개선된 버전
	var contentBuilder strings.Builder
	contentSelectors := []string{
		".se-main-container",
		".post_content",
		".se-component.se-text.se-section",
		".sect_dsc",
		".post_ct",
		"#content-area .post_content",
		".se-module-text",
		".pcol1 .post_content",
	}

	contentFound := false
	for _, sel := range contentSelectors {
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			// 불필요한 요소 제거
			s.Find("img, .se-sticker, .se-module-oglink, .se-map-container, .se-file-block, script, style").Remove()

			// 텍스트 추출
			text := utils.CleanText(s.Text())
			if text != "" && len(text) > 10 { // 최소 길이 체크
				contentBuilder.WriteString(text)
				contentBuilder.WriteString("\n")
				contentFound = true
			}
		})

		if contentFound {
			break
		}
	}

	blogPost.Content = utils.CleanText(contentBuilder.String())

	// 댓글 추출
	var comments []BlogComment
	commentSelectors := []string{
		".comment_area .comment_item",
		"._commentWrapper .comment_row",
		".comment_list .comment",
		".cmt_area .cmt_item",
	}

	for _, sel := range commentSelectors {
		doc.Find(sel).Each(func(i int, s *goquery.Selection) {
			commentContent := utils.CleanText(s.Find(".comment_text, .text_comment, .cmt_text").First().Text())
			commentWriter := utils.CleanText(s.Find(".comment_nick, .author_name, .cmt_nick").First().Text())
			commentDate := utils.CleanText(s.Find(".comment_date, .date, .cmt_date").First().Text())

			if commentContent != "" {
				comments = append(comments, BlogComment{
					ID:        fmt.Sprintf("%d", i+1),
					Content:   commentContent,
					Writer:    commentWriter,
					WriteDate: commentDate,
				})
			}
		})

		if len(comments) > 0 {
			break
		}
	}

	blogPost.Comments = comments

	// 디버그 정보 출력
	log.Printf("📝 게시글 처리 완료 - ID: %s", articleID)
	log.Printf("   제목: %s", blogPost.Title)
	log.Printf("   작성자: %s", blogPost.Writer)
	log.Printf("   작성일: %s", blogPost.WriteDate)
	log.Printf("   내용 길이: %d글자", len(blogPost.Content))
	log.Printf("   댓글 수: %d개", len(comments))

	// 필수 정보가 없으면 에러 반환
	if blogPost.Title == "" && blogPost.Content == "" {
		return blogPost, fmt.Errorf("게시글 정보를 추출할 수 없습니다 (제목과 내용 모두 비어있음)")
	}

	return blogPost, nil
}
