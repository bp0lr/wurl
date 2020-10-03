# wurl

this tool will check Concurrently the HTTP/S status code for each link agains a status code list.

Useful when you got a lot of links from waybackmachine and want to clean up the list a bit.

Wurl has multiple options to make your life easier.

# options
```
 ▶ wurl -h 

Usage of wurl:
  -c, --concurrency int       Concurrency level (default 20)
  -m, --finger-print string   regex for a specific string on response
  -f, --follow-redirect       Follow redirects (Default: false)
  -H, --header stringArray    Add custom Headers to the request
  -o, --output string         Output file to save the results to
  -p, --proxy string          Add a HTTP proxy
  -u, --random-agent          Set a random User Agent
  -s, --status-code string    List valid status codes (default 200)
  -t, --test                  Test http && https for a single url
  -v, --verbose               Display extra info about what is going on
pflag: help requested
exit status 2

```

## Install
```
 ▶ go get -u github.com/bp0lr/wurl
```


## usage examples
```
 ▶ cat urls.txt | wurl -c 50
-> get each link and return just working ones. (status code 200).
```
```
 ▶ cat urls.txt | wurl -c 50 -f
-> get each link following redirection and return working ones. (status code 200 / 302).
```
```
▶ cat urls.txt | wurl -c 50 -H ""Cookie: mycookie" -s "200, 500, 302"
-> get each link using my cookie and return status codes [200, 500, 302]
```
```
▶ cat urls.txt | wurl -c 50 -H ""Cookie: mycookie" -s "200, 500, 302" -p "http://192.168.0.150:8080"
-> get each link using my cookie and a proxy server (burp for example) and return status codes [200, 500, 302]
```
```
▶ cat urls.txt | wurl -c 50 -H ""Cookie: mycookie" -s "200, 500, 302" -p "http://192.168.0.150:8080" -f "paypal"
-> get each link using my cookie and a proxy server (burp for example).
return status codes [200, 500, 302] and results containing the word "paypal" in his body.
```
```
▶ cat urls.txt | wurl -c 50 -H ""Cookie: mycookie" -s "200, 500, 302" -p "http://192.168.0.150:8080" -f "paypal" --random-agent
-> get each link using my cookie, a random-agent for each query and a proxy server (burp for example).
return status codes [200, 500, 302] and results containing the word "paypal" in his body.
```
```
▶ cat urls.txt | wurl -c 50  -s "200" --random-agent -test
-> send a HTTP and HTTPS query for each link using a random-agent and return just working ones.
return status codes [200].
```
