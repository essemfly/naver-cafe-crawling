# 네이버 카페 크롤러

## 📝 프로젝트 설명
이 프로젝트는 네이버 카페의 게시글을 크롤링하는 Go 언어 기반의 크롤러입니다.

## 🚀 시작하기

### 필수 조건
- Go 1.16 이상
- 네이버 카페 로그인 쿠키

### 설치
```bash
git clone https://github.com/yourusername/naverCafeCrawler.git
cd naverCafeCrawler
go mod download
```

### 환경 변수 설정
`.env` 파일을 생성하고 다음 내용을 입력하세요:
```env
NAVER_CAFE_ID=your_cafe_id
NAVER_BOARD_ID=your_board_id
NAVER_COOKIE=your_cookie
```

### 실행
```bash
go run main.go
```

## 💾 결과 저장
크롤링 결과는 `output` 폴더에 JSON 파일로 저장됩니다.

### 파일명 형식
- 전체 결과: `cafe_{카페ID}_board_{게시판ID}_{타임스탬프}_full.json`
- 페이지별 결과: `cafe_{카페ID}_board_{게시판ID}_{타임스탬프}_page_{페이지번호}.json`

### JSON 구조
```json
[
  {
    "id": "게시글ID",
    "title": "제목",
    "writer": "작성자",
    "write_date": "작성일시",
    "read_count": "조회수",
    "comment_count": "댓글수",
    "like_count": "좋아요수",
    "content": "게시글 내용 (HTML 형식)",
    "comments": [
      {
        "writer": "댓글 작성자",
        "content": "댓글 내용 (HTML 형식)",
        "write_date": "댓글 작성일시"
      }
    ]
  }
]
```

## 🔧 HTML 파싱 필요사항
현재 크롤러는 게시글 내용과 댓글을 HTML 형식으로 가져옵니다. 실제 사용을 위해서는 다음 작업이 필요합니다:

1. HTML 파싱 라이브러리 추가
   - `golang.org/x/net/html` 또는 `github.com/PuerkitoBio/goquery` 등의 라이브러리 사용 권장

2. 파싱 기능 구현
   - HTML 태그 제거
   - 특수 문자 처리
   - 이미지 URL 추출
   - 링크 처리
   - 이모지/이모티콘 처리

3. 파싱된 결과 저장
   - 원본 HTML과 파싱된 텍스트를 함께 저장
   - 이미지 URL 목록 별도 저장
   - 링크 정보 별도 저장

## ⚠️ 주의사항
- 네이버 카페의 이용약관을 준수하여 사용하세요.
- 과도한 요청은 IP 차단의 원인이 될 수 있습니다.
- 크롤링한 데이터의 저작권을 존중하세요.

## 📄 라이선스
이 프로젝트는 MIT 라이선스 하에 배포됩니다. 