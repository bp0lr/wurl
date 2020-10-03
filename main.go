//
// @bp0lr - 10/02/2020
//

package main

import (

	"fmt"
	"bufio"	
	"os"
	"sync"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
	"strings"
	"math/rand"
	"strconv"
	"regexp"
	"io/ioutil"

	flag 		"github.com/spf13/pflag"
				"github.com/forestgiant/sliceutil"
	)
	
	var (
		concurrencyArg 			int
		HeaderArg 				[]string
		statusListArg 			string
		proxyArg 				string
		fingerPrintArg 			string
		outputFileArg 			string
		verboseArg 				bool
		followRedirectArg 		bool
		useRandomAgentArg 		bool
		testHTTPArg 			bool		
	)


func main() {

	flag.StringArrayVarP(&HeaderArg, "header", "H", nil, "Add custom Headers to the request")
	flag.IntVarP(&concurrencyArg, "concurrency", "c", 20, "Concurrency level")
	flag.StringVarP(&statusListArg, "status-code", "s", "", "List valid status codes.")
	flag.BoolVarP(&verboseArg, "verbose", "v", false, "Display extra info about what is going on")
	flag.BoolVarP(&followRedirectArg, "follow-redirect", "f", false, "Follow redirects (Default: false)")
	flag.StringVarP(&proxyArg, "proxy", "p", "", "Add a HTTP proxy")
	flag.BoolVarP(&useRandomAgentArg, "random-agent", "u", false, "Set a random User Agent")
	flag.StringVarP(&fingerPrintArg, "finger-print", "m", "", "regex for a specific string on response")
	flag.StringVarP(&outputFileArg, "output", "o", "", "Output file to save the results to")
	flag.BoolVarP(&testHTTPArg, "test", "t", false, "Test http && https for a single url")
	
	flag.Parse()

	//concurrency
	concurrency := 20
	if concurrencyArg > 0 {
		concurrency = concurrencyArg
	}

	//status code
	status := strings.Split(statusListArg, ",")
	if(len(status[0]) < 1){
		status = status[:0]
		status = append(status, "200")
	}

	client := newClient(proxyArg, followRedirectArg)
	
	jobs := make(chan string)
	var wg sync.WaitGroup

	var outputFile *os.File
	var err0 error
	if(outputFileArg != ""){
		outputFile, err0 = os.OpenFile(outputFileArg, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err0 != nil {
			fmt.Printf("cannot write Volume Data %s.dat: %s", outputFileArg, err0.Error())
			return
		}
		defer outputFile.Close()
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			for raw := range jobs {
				
				u, err := url.ParseRequestURI(raw)
				if err != nil {
					if(verboseArg){
						fmt.Printf("[-] Invalid url: %s\n", raw)
					}
					continue
				}

				if(testHTTPArg){
					if strings.HasPrefix(u.String(), "http://") {
						processRequest(u, client, status, outputFile)
						alt, _:= url.ParseRequestURI(strings.Replace(u.String(), "http:", "https:", 1))
						processRequest(alt, client, status, outputFile)
					}else{
						processRequest(u, client, status, outputFile)
						alt, _:= url.ParseRequestURI(strings.Replace(u.String(), "https:", "http:", 1))
						processRequest(alt, client, status, outputFile)
					}					
				}else{
					processRequest(u, client, status, outputFile)
				}
					
				
			}
			wg.Done()
		}()
	}

	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		jobs <- sc.Text()
	}
	close(jobs)
	wg.Wait()
}

func processRequest(u *url.URL, client *http.Client, status []string, outputFile *os.File){

	if(verboseArg){
		fmt.Printf("[+] Testing: %v\n", u.String())
	}
	
	req, err := http.NewRequest("GET", u.String(), nil)
	
	if err != nil {
		if(verboseArg){
			fmt.Printf("[-] Error: %v\n", err)
		}
		return
	}
	
	if useRandomAgentArg {
		req.Header.Set("User-Agent", getUserAgent())
	}else{
		req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; wurl/1.0)")
	}
	
	// add headers to the request
	for _, h := range HeaderArg {
		parts := strings.SplitN(h, ":", 2)

		if len(parts) != 2 {
			continue
		}
		req.Header.Set(parts[0], parts[1])
	}

	// send the request
	resp, err := client.Do(req)
	if err != nil {
		if(verboseArg){
			fmt.Printf("[-] Error: %v\n", err)
		}
		return
	}
	defer resp.Body.Close()

	if(sliceutil.Contains(status, strconv.Itoa(resp.StatusCode))){
		if(verboseArg){
				fmt.Printf("[+] %v [%v]\n", u.String(), resp.StatusCode)
		}else{						
			if fingerPrintArg != "" {
				data, _ := ioutil.ReadAll(resp.Body)
				var re, _ = regexp.Compile(fingerPrintArg)
				
				if re.MatchString(string(data)) == true {
					if(outputFileArg != ""){
						outputFile.WriteString(u.String() + "\n")
					}else{
						fmt.Printf("%v\n", u.String())
					}
				}
			}else{
				if(outputFileArg != ""){
					outputFile.WriteString(u.String() + "\n")
				}else{
					fmt.Printf("%v\n", u.String())
				}
			}												
		}
	}else{
		if(verboseArg){
			fmt.Printf("[-] %v [%v]\n", u.String(), resp.StatusCode)
		}
	}

}

func newClient(proxy string, followRedirect bool) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:		30,
		IdleConnTimeout:	time.Second,
		TLSClientConfig:	&tls.Config{InsecureSkipVerify: true},
		DialContext:		(&net.Dialer{
		Timeout:			time.Second * 5,
		}).DialContext,
	}

	if proxy != "" {
		if p, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(p)
		}
	}

	client := &http.Client{
		Transport:		tr,
		Timeout:		time.Second * 5,
	}

	if !followRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

func getUserAgent() string {
	payload := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:66.0) Gecko/20100101 Firefox/66.0",
  		"Mozilla/5.0 (Windows NT 6.2; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.106 Safari/537.36",
  		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.1 Safari/605.1.15",
  		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36",
  		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:67.0) Gecko/20100101 Firefox/67.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 8_4_1 like Mac OS X) AppleWebKit/600.1.4 (KHTML, like Gecko) Version/8.0 Mobile/12H321 Safari/600.1.4",
		"Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (iPad; CPU OS 7_1_2 like Mac OS X) AppleWebKit/537.51.2 (KHTML, like Gecko) Version/7.0 Mobile/11D257 Safari/9537.53",
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; Trident/6.0)",
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(payload))
	
	pick := payload[randomIndex]

	return pick
}