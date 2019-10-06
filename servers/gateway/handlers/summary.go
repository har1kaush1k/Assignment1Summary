package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
}

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	/*TODO: add code and additional functions to do the following:
	- Add an HTTP header to the response with the name
	 `Access-Control-Allow-Origin` and a value of `*`. This will
	  allow cross-origin AJAX requests to your server.
	- Get the `url` query string parameter value from the request.
	  If not supplied, respond with an http.StatusBadRequest error.
	- Call fetchHTML() to fetch the requested URL. See comments in that
	  function for more details.
	- Call extractSummary() to extract the page summary meta-data,
	  as directed in the assignment. See comments in that function
	  for more details
	- Close the response HTML stream so that you don't leak resources.
	- Finally, respond with a JSON-encoded version of the PageSummary
	  struct. That way the client can easily parse the JSON back into
	  an object. Remember to tell the client that the response content
	  type is JSON.

	Helpful Links:
	https://golang.org/pkg/net/http/#Request.FormValue
	https://golang.org/pkg/net/http/#Error
	https://golang.org/pkg/encoding/json/#NewEncoder
	*/
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		http.Error(w, "No query found in the requested url", http.StatusBadRequest)
	}
	response, err := fetchHTML(url)
	if err != nil {
		log.Fatal("could not fetch url ")
	}
	targetSummary, err := extractSummary(url, response)
	if err != nil {
		log.Fatal("error extracting summary")
	}
	jsonError := json.NewEncoder(w).Encode(targetSummary)
	if jsonError != nil {
		log.Fatal("Error encoding the summary to json")
	}

}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	/*TODO: Do an HTTP GET for the page URL. If the response status
	code is >= 400, return a nil stream and an error. If the response
	content type does not indicate that the content is a web page, return
	a nil stream and an error. Otherwise return the response body and
	no (nil) error.

	To test your implementation of this function, run the TestFetchHTML
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestFetchHTML

	Helpful Links:
	https://golang.org/pkg/net/http/#Get
	*/
	resp, err := http.Get(pageURL)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("response status code was %d", resp.StatusCode)
	}

	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("response content type was %s, not text/html", ctype)
	}

	return resp.Body, err
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	/*TODO: tokenize the `htmlStream` and extract the page summary meta-data
	according to the assignment description.

	To test your implementation of this function, run the TestExtractSummary
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestExtractSummary
	Helpful Links:
	https://drstearns.github.io/tutorials/tokenizing/
	http://ogp.me/
	https://developers.facebook.com/docs/reference/opengraph/
	https://golang.org/pkg/net/url/#URL.ResolveReference
	*/
	resSummary := &PageSummary{}

	tokenizer := html.NewTokenizer(htmlStream)

	for {
		tokenType := tokenizer.Next()
		token := tokenizer.Token()
		if tokenType == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			if token.Data == "meta" {
				property := getTargetAttr(token, "property")
				name := getTargetAttr(token, "name")
				content := getTargetAttr(token, "content")
				switch property {
				case "og:type":
					resSummary.Type = content
				case "og:url":
					resSummary.URL = content
				case "og:title":
					resSummary.Title = content
				case "og:site_name":
					resSummary.SiteName = content
				}
				if property == "og:description" {
					resSummary.Description = content
				} else if name == "description" && resSummary.Description == "" {
					resSummary.Description = content
				}

				if name == "author" {
					resSummary.Author = content
				}

				if name == "keywords" {
					arr := strings.Split(content, ",")
					for i := range arr {
						arr[i] = strings.TrimSpace(arr[i])
					}
					resSummary.Keywords = arr
				}

				if strings.HasPrefix(property, "og:image") {
					if strings.HasPrefix(property, "og:image:") {
						recentImage := resSummary.Images[len(resSummary.Images)-1]
						switch property {
						case "og:image:secure_url":
							recentImage.SecureURL = getAbsoluteURL(pageURL, content)
						case "og:image:alt":
							recentImage.Alt = content
						case "og:image:type":
							recentImage.Type = content
						case "og:image:width":
							recentImage.Width, _ = strconv.Atoi(content)
						case "og:image:height":
							recentImage.Height, _ = strconv.Atoi(content)
						}
					} else {
						newImg := &PreviewImage{}
						if !strings.HasPrefix(content, "http://") {
							content= getAbsoluteURL(pageURL, content)
						}
						newImg.URL = content
						resSummary.Images = append(
							resSummary.Images, newImg)
					}
				}
			}
			if token.Data == "title" && resSummary.Title == "" {
				next := tokenizer.Next()
				if next == html.TextToken {
					resSummary.Title = tokenizer.Token().Data
				}
			}

			if token.Data == "link" {
				icon := &PreviewImage{}
				rel := getTargetAttr(token, "rel")
				if rel == "icon" {
					icon.URL = getTargetAttr(token, "href")
					icon.Type = getTargetAttr(token, "type")
					icon.Alt  = getTargetAttr(token, "alt")
					sizes  := getTargetAttr(token, "sizes")

					if !strings.HasPrefix(icon.URL, "http://") {
						relative, _ := url.Parse(icon.URL)
						absolutePattern, _ := url.Parse(pageURL)
						icon.URL = absolutePattern.ResolveReference(relative).String()
					}
					if sizes != "any" && sizes != "" {
						sizeSlice := strings.Split(sizes, "x")
						icon.Height, _ = strconv.Atoi(sizeSlice[0])
						icon.Width, _ = strconv.Atoi(sizeSlice[1])
					}
				}
				resSummary.Icon = icon
			}
		}

		if tokenType == html.EndTagToken && token.Data == "head" {
			break
		}
	}
	return resSummary, nil
}


func getTargetAttr(token html.Token, target string) string {
	for _, a := range token.Attr {
		if a.Key == target {
			return a.Val
		}
	}
	return ""
}

func getAbsoluteURL(absoluteBase string, relative string) string {
	absoluteURL, _ := url.Parse(absoluteBase)
	relativeURL, _ := url.Parse(relative)
	return absoluteURL.ResolveReference(relativeURL).String()
}

