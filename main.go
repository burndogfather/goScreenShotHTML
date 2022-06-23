package main
import (
	"net/http"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

//요청이 들어오면 실행되는 함수
func requestHandler(res http.ResponseWriter, req *http.Request) {
	req.ParseForm() //form데이터 수집
	
	
	postdata := req.PostForm //post데이터만 담기
	
	fmt.Println(postdata)
	if postdata["url"] != nil{ //post데이터에서 url이라는 값찾기
		url, _ := postdata["target_url"]
		
		fmt.Println(url)
	}
}

func main() {
	
	//8000번 포트로 http 서버열기
	//nginx연결됨 (https://git.coco.sqs.kr/proxy-8000)
	err := http.ListenAndServe(":8000", http.HandlerFunc(requestHandler))
	if err != nil {
		//http 서버 실행실패시 에러처리
		fmt.Println("Failed to ListenAndServe : ", err)
	}
	
	
	taskCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 15*time.Second)
	defer cancel()
	var pdfBuffer []byte
	if err := chromedp.Run(taskCtx, pdfGrabber("https://www.naver.com", "body", &pdfBuffer)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("naver.pdf", pdfBuffer, 0644); err != nil {
		log.Fatal(err)
	}
}

func pdfGrabber(url string, sel string, res *[]byte) chromedp.Tasks {

	start := time.Now()
	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			fmt.Printf("\nDuration: %f secs\n", time.Since(start).Seconds())
			return nil
		}),
	}
}