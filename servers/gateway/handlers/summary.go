package handlers

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/net/html"
	"io"
	"net/http"
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
	fmt.Println(resp.Status)
	if resp.StatusCode >= 400 {
		fmt.Println(err)
		return nil, errors.New("Status code not 200 OK")
	}
	if resp.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		return nil, errors.New("Content-Type is not text/html")
	}
	defer resp.Body.Close()
	return resp.Body, nil
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

	tokenizer := html.NewTokenizer(htmlStream)
	var pageSummary PageSummary
	structMap := map[string]string{}
	imageMap := map[string]string{}
	imageField := ""
	shouldLoop := true
	for shouldLoop {
		tokenType := tokenizer.Next()
		//fmt.Println(tokenType)
		//fmt.Println(tokenizer.Token())
		//fmt.Println(html.StartTagToken)
		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				//end of the file, break out of the loop
				shouldLoop = false
			}
			return nil, tokenizer.Err()
		//case html.StartTagToken:
		case html.StartTagToken:
			//fmt.Println("found StartTagToken")
			token := tokenizer.Token()
			if token.Data == "meta" {
				fieldName := "Type"
				fmt.Println("found meta")
				for _, element := range token.Attr {
					// fmt.Println(element)
					if element.Key == "property" {
						switch element.Val {
						case "og:type":
							fieldName = "Type"
						case "og:url":
							fieldName = "URL"
						case "og:title":
							fieldName = "Title"
						case "og:site_name":
							fieldName = "SiteName"
						case "og:description":
							fieldName = "Description"
						case "og:image":
							fieldName = "Images"
						case "og:image:url":
							imageField = "URL"
						case "og:image:secure_url":
							imageField = "SecureURL"
						case "og:image:type":
							imageField = "Type"
						case "og:image:width":
							imageField = "Width"
						case "og:image:height":
							imageField = "Height"
						case "og:image:alt":
							imageField = "Alt"

						}
					} else if element.Key == "name" {
						switch element.Val {
						case "description":
							fieldName = "Description"
						case "author":
							fieldName = "Author"
						case "keywords":
							fieldName = "Keywords"
						}
					} else if element.Key == "content" {
						structMap[fieldName] = element.Val
						imageMap[imageField] = element.Val
					}
				}
			}
			if token.Data == "link" {

			}
		}
		pageSummary = *initializeSummary(&structMap, &imageMap)
		fmt.Println(structMap)

	}
	return &pageSummary, nil

	//return nil, nil
}

func initializeSummary(summaryMap *map[string]string, imageMap *map[string]string) *PageSummary  {
	res := PageSummary{}
	for key, val := range *summaryMap{
		switch key {
		case "Type":
			res.Type = val
		case "URL":
			res.URL = val
		case "Title":
			res.Title = val
		case "SiteName":
			res.SiteName = val
		case "Description":
			res.Description = val
		case "Author":
			res.Author = val
		case "Keywords":
			arr := strings.Split(val, ",")
			res.Keywords = arr
		case "Icon":
			icon := PreviewImage{}


			res.Icon = val
		case "Images":
			imageArr := make([]PreviewImage, 1)
			for k, v := range *imageMap {
				switch k {
				case "URL":
				case "SecureURL":
				case "Type":
				case "Width":
				case "Height":
				case "Alt":
				}
			}
			res.Images = val
		}
	}
	return &res
}

