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
	//FORM > POST 데이터 가져오기
	req.ParseForm()
	postdata := req.PostForm
	
	//POST 데이터에서 url이라는 값을 찾아서 String을 벗기기(?)
	if ( postdata["url"] != nil && postdata["element"] != nil){ 
		var url string
		var element string
		for _, url := range postdata["url"] {
			fmt.Println(url)
		}
		
		for _, element := range postdata["element"] {
			fmt.Println(element)
		}
			
		//여기서부터 Chromedp설정
		taskCtx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		defer cancel()
		
		//최대 대기시간은 15초
		taskCtx, cancel = context.WithTimeout(taskCtx, 15*time.Second)
		defer cancel()
		
		//사이트 캡쳐
		var pdfBuffer []byte
		if err := chromedp.Run(taskCtx, pdfGrabber(url, element, &pdfBuffer)); err != nil {
			log.Fatal(err)
		}
		
		//파일로 저장
		if err := ioutil.WriteFile("naver.pdf", pdfBuffer, 0644); err != nil {
			log.Fatal(err)
		}
		
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
	
}

func pdfGrabber(url string, sel string, res *[]byte) chromedp.Tasks {

	start := time.Now()
	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
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