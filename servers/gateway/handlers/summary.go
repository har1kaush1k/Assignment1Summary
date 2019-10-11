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


// This function handles the requests for the page summary API
// It takes in a URL string parameter and returns a JSON-encoeded
// struct that contains the page summary meta-data
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
	// Allowing cross-origin AJAX requests to the server
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")

	// Grabbiing the URL and throwing an error if not supplied
	url := r.URL.Query().Get("url")
	if len(url) == 0 {
		http.Error(w, "No query found in the requested url", http.StatusBadRequest)
	}

	// Fetching URL
	response, err := fetchHTML(url)
	if err != nil {
		log.Fatal("could not fetch url ")
	}

	// Extracting page summary meta-data, throwing error if nil
	// Closing within extractSumary method
	targetSummary, err := extractSummary(url, response)
	if err != nil {
		log.Fatal("error extracting summary")
	}

	// Responded with JSON-encoded version of PageSummary struct
	// Throws an error if it returns nil
	jsonError := json.NewEncoder(w).Encode(targetSummary)
	if jsonError != nil {
		log.Fatal("Error encoding the summary to json")
	}

}


// Fetching the 'pageURL' and returning the body stream of the page
// If response status is >= 400 or if content type does not 
// indicate that the content is a web page, return a nil stream and an error
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

	// Getting the URL
	resp, err := http.Get(pageURL)


	// Return status coded if status is less than 400
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("response status code was %d", resp.StatusCode)
	}

	// Get and check the content type from the header
	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("response content type was %s, not text/html", ctype)
	}

	// Closing the response body
	// This code wasn't in the orginal plan we submitted
	// but I think we are missing it:
	// defer resp.Body.Close()

	return resp.Body, err
}


 // Handles the tokinization of the 'htmlStream' and sends that meta-data
 // to fill the PageSummary struct
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

	// Creating a new Tokenizer with the htmlStream
	tokenizer := html.NewTokenizer(htmlStream)

	//Looping through till finding the first
	//StartTagToken or SelfClosingTagToken and its data
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
			//SelfClosingTagToken or StartTagToken have been found
			if token.Data == "meta" {
				
				property := getTargetAttr(token, "property")
				name := getTargetAttr(token, "name")
				content := getTargetAttr(token, "content")

				// Assiging variables based on meta property tag
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

				// Spliting keywords into an array
				// And storing that array
				if name == "keywords" {
					arr := strings.Split(content, ",")
					for i := range arr {
						arr[i] = strings.TrimSpace(arr[i])
					}
					resSummary.Keywords = arr
				}


				// If it is a image tag, assign variables accordingly
				// to allow for preview of image
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

			// Assigning the title tag 
			if token.Data == "title" && resSummary.Title == "" {
				next := tokenizer.Next()
				if next == html.TextToken {
					resSummary.Title = tokenizer.Token().Data
				}
			}

			// If the Token Data is a link,
			// it shows a preview of the image
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

		//Only looking at tokens within HEAD html tags
		if tokenType == html.EndTagToken && token.Data == "head" {
			break
		}
	}
	return resSummary, nil
}


// Returning the target attribute of a Token
func getTargetAttr(token html.Token, target string) string {
	for _, a := range token.Attr {
		if a.Key == target {
			return a.Val
		}
	}
	return ""
}


// Converting from relative URL to return a absoluteURL
func getAbsoluteURL(absoluteBase string, relative string) string {
	absoluteURL, _ := url.Parse(absoluteBase)
	relativeURL, _ := url.Parse(relative)
	return absoluteURL.ResolveReference(relativeURL).String()
}

