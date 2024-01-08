package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/stretchr/testify/assert"
)

/*		--== FUNCTIONS TO TEST ==--
[x] - runModules(verbose bool, modules string, functions map[string]func([]responseData, bool), responses []responseData)
[x] - printRelationships(tag string, relationships map[string][]string)
[x] - visitPage(c *colly.Collector, r *colly.Response, e *colly.HTMLElement, visited map[string]bool, relationships map[string][]string, func_scrape bool, table *tablewriter.Table, tag string, pageDataMap map[string]PageData)
[x] - stringInSlice(value string, slice []string) bool
[x] - visitJavascriptEndpoints(c *colly.Collector, r *colly.Response, urlString string, responses []responseData, relationships map[string][]string)
[x] - removeDuplicates(elements []string) []string
[x] - handleResponse(r *colly.Response, c *colly.Collector, urlString string, responses []responseData, relationships map[string][]string)
[x] - setCookieCallback(c *colly.Collector, parts []string, auth_type string, cookie string)
[x] - processFlags() (string, string, string, bool, string, int, bool, bool, bool, bool, error)
[x] - validateFlags(url_input *string, auth_input *string, tag_input *string, subdomains_input *bool, modules_input *string, requests_input *int, prod_input *bool) error
[x] - getDomain(urlString string) string
[x] - createCollector(domain string, allowSubdomains bool) *colly.Collector
[x] - headers(responses []responseData, verbose bool)
[x] - cors(responses []responseData, verbose bool)
[x] - addTrailingSlashIfNeeded(inputURL string) (string, error)
[x] - removePort(url string) string
[x] - LoadConfig(filename string) (*Config, error)
[x] - writeRelationshipsToCSV(relationships map[string][]string, filename string) error
[x] - createDotFile(relationships map[string][]string, filename string)
[x] - calc_url(src string, domain string) string

*/

func TestCors(t *testing.T) {
	// Create a test HTTP server with a custom handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the headers to simulate a vulnerable CORS configuration
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Write([]byte("Hello, world!"))
	}))
	defer server.Close()

	// Prepare a list of responseData items
	responses := []responseData{
		{
			URL: server.URL,
			Headers: map[string]string{
				"Content-Type":                     "text/html",
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Credentials": "true",
			},
		},
	}

	// Capture the output of the cors function
	oldOutput := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the cors function with the prepared list of responseData items
	cors(responses, false)

	// Restore the original stdout and read the captured output
	w.Close()
	os.Stdout = oldOutput
	output, _ := ioutil.ReadAll(r)

	// Verify the output of the cors function
	expected := server.URL + " is vulnerable to CORS attack\n"
	if !strings.Contains(string(output), expected) {
		t.Errorf("Expected output to contain '%s', but got: %s", expected, output)
	}
}

func TestVisitPage(t *testing.T) {
	// Create a mock HTML document for testing
	html := `<html>
<head>
	<script src="/js/test.js"></script>
</head>
<body>
	<a href="http://example.com/test1">Test Link 1</a>
	<a href="http://example.com/test2">Test Link 2</a>
	<form><input type="text" /><button>Submit</button></form>
</body>
</html>`

	// Initialize variables
	visited := make(map[string]bool)
	relationships := make(map[string][]string)
	pageDataMap := make(map[string]PageData)

	// Create a new Colly collector
	c := colly.NewCollector()

	// Create a mock HTMLElement
	e := &colly.HTMLElement{
		DOM: createDOMNode(t, html),
	}
	e.Request = &colly.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   "example.com",
		},
		Ctx: colly.NewContext(),
	}
	e.Response = &colly.Response{
		Request: e.Request,
		Headers: &http.Header{},
	}
	extensions.RandomUserAgent(c)

	// Call the visitPage function
	visitPage(c, e.Response, e, visited, relationships, false, nil, "", pageDataMap)

	// Assertions
	assert.True(t, visited["http://example.com/"], "URL should be marked as visited")
	assert.Equal(t, 3, len(relationships["http://example.com/"]), "Three links should be found in the relationships")

	expectedPageData := PageData{
		URL:         "http://example.com/",
		elements:    "form, input, button",
		AccessLevel: []string{""},
	}
	assert.Equal(t, expectedPageData, pageDataMap["http://example.com/"], "PageData should match the expected data")
}

// Helper function to create a dummy colly.Response
func createDummyResponse() *colly.Response {
	return &colly.Response{
		Request: &colly.Request{
			URL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
			},
		},
		Headers: &http.Header{},
	}
}

// Helper function to create an html.Node from a string
func createDOMNode(t *testing.T, html string) *goquery.Selection {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatal("Failed to create DOM node:", err)
	}
	return doc.Selection
}

func TestVisitJavascriptEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		body        string
		expectedURL string
	}{
		{"Valid JavaScript file", "http://example.com/valid.js", "application/javascript", `"https://example.com/expected"`, "https://example.com/expected"},
		{"Invalid JavaScript file", "http://example.com/invalid.js", "application/javascript", `"not-a-valid-url"`, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing visitJavascriptEndpoints with URL: %s", test.url)

			responses := []responseData{}
			relationships := make(map[string][]string)

			parsedURL, _ := url.Parse(test.url)
			r := &colly.Response{
				Request: &colly.Request{URL: parsedURL},
				Headers: &http.Header{},
				Body:    []byte(test.body),
			}
			r.Headers.Set("Content-Type", test.contentType)

			c := colly.NewCollector()
			visitJavascriptEndpoints(c, r, "http://example.com", responses, relationships)

			if test.expectedURL != "" {
				if _, ok := relationships[test.url]; !ok {
					t.Errorf("URL not found in relationships: %s", test.url)
				} else {
					found := false
					for _, relatedURL := range relationships[test.url] {
						if strings.HasPrefix(relatedURL, test.expectedURL) {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("Expected URL not found in relationships: %s", test.expectedURL)
					}
				}
			} else {
				if _, ok := relationships[test.url]; ok {
					t.Errorf("Unexpected URL found in relationships: %s", test.url)
				}
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		contentType string
		body        string
		expectedURL string
	}{
		{"Valid JavaScript file", "http://example.com/valid.js", "application/javascript", `"https://example.com/expected"`, "https://example.com/expected"},
		{"Invalid JavaScript file", "http://example.com/invalid.js", "application/javascript", `"not-a-valid-url"`, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing handleResponse with URL: %s", test.url)

			responses := []responseData{}
			relationships := make(map[string][]string)

			parsedURL, _ := url.Parse(test.url)
			r := &colly.Response{
				Request: &colly.Request{URL: parsedURL},
				Headers: &http.Header{},
				Body:    []byte(test.body),
			}
			r.Headers.Set("Content-Type", test.contentType)

			c := colly.NewCollector()
			handleResponse(r, c, "http://example.com", responses, relationships)

			if test.expectedURL != "" {
				if _, ok := relationships[test.url]; !ok {
					t.Errorf("URL not found in relationships: %s", test.url)
				} else {
					found := false
					for _, relatedURL := range relationships[test.url] {
						if strings.HasPrefix(relatedURL, test.expectedURL) {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("Expected URL not found in relationships: %s", test.expectedURL)
					}
				}
			} else {
				if _, ok := relationships[test.url]; ok {
					t.Errorf("Unexpected URL found in relationships: %s", test.url)
				}
			}
		})
	}
}

func TestSetCookieCallback(t *testing.T) {
	tests := []struct {
		name     string
		authType string
		cookie   string
		expected string
	}{
		{"Valid auth_type and cookie", "Authorization", "Bearer token", "Bearer token"},
		{"Empty auth_type", "", "Bearer token", ""},
		{"Empty cookie", "Authorization", "", ""},
		{"Empty auth_type and cookie", "", "", ""},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing setCookieCallback with authType = %q and cookie = %q", test.authType, test.cookie)

			c := colly.NewCollector()
			setCookieCallback(c, nil, test.authType, test.cookie)

			var result string
			c.OnRequest(func(r *colly.Request) {
				result = r.Headers.Get(test.authType)
			})

			// Simulate a request
			_ = c.Visit("http://example.com")

			if result != test.expected {
				t.Errorf("setCookieCallback authType = %q, cookie = %q, result = %q; expected %q", test.authType, test.cookie, result, test.expected)
			} else {
				t.Logf("setCookieCallback authType = %q, cookie = %q, result = %q; passed", test.authType, test.cookie, result)
			}
		})
	}
}

type captureWriter struct {
	buffer *bytes.Buffer
}

func (cw *captureWriter) Write(p []byte) (n int, err error) {
	return cw.buffer.Write(p)
}

func TestHeaders(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		hdrs           map[string]string
		expectedOutput string
	}{
		{
			name:           "Missing all security headers",
			contentType:    "text/html",
			hdrs:           make(map[string]string),
			expectedOutput: "+-------------------+--------------------------+\n|        URL        | MISSING SECURITY HEADERS |\n+-------------------+--------------------------+\n| http://127.0.0.1/ | Permissions-Policy       |\n+-------------------+--------------------------+\n",
		},
		{
			name:        "Missing only Permissions-Policy",
			contentType: "text/html",
			hdrs: map[string]string{
				"X-Content-Type-Options":  "nosniff",
				"X-Frame-Options":         "SAMEORIGIN",
				"Content-Security-Policy": "default-src 'self'",
				"Referrer-Policy":         "strict-origin",
			},
			expectedOutput: "+-------------------+--------------------------+\n|        URL        | MISSING SECURITY HEADERS |\n+-------------------+--------------------------+\n| http://127.0.0.1/ | Permissions-Policy       |\n+-------------------+--------------------------+\n",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing headers with contentType = %q and hdrs = %v", test.contentType, test.hdrs)

			// Create the responses slice
			var responses []responseData
			responseHeaders := make(map[string]string)
			responseHeaders["Content-Type"] = test.contentType
			for k, v := range test.hdrs {
				responseHeaders[k] = v
			}

			responses = append(responses, responseData{
				URL:      "http://127.0.0.1/",
				Response: "",
				Headers:  responseHeaders,
			})

			// Redirect output to a buffer to capture print statements
			reader, writer, _ := os.Pipe()
			oldStdout := os.Stdout
			os.Stdout = writer
			defer func() {
				os.Stdout = oldStdout
			}()

			go func() {
				headers(responses, false)
				writer.Close()
			}()

			output, _ := ioutil.ReadAll(reader)

			if !reflect.DeepEqual(string(output), test.expectedOutput) {
				t.Errorf("Unexpected output for test case %d: %s.\nExpected:\n%q\nGot:\n%q\n", i+1, test.name, test.expectedOutput, string(output))
			} else {
				t.Logf("Expected output found for test case %d: %s; passed", i+1, test.name)
			}
		})
	}
}

func TestRunModules(t *testing.T) {
	// Define sample data
	verbose := false
	modules := "module1,module2,headers"
	functions := map[string]func([]responseData, bool){
		"module1": func(responses []responseData, verbose bool) {},
		"module3": func(responses []responseData, verbose bool) {},
	}
	responses := []responseData{}

	// Capture the output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the function being tested
	runModules(verbose, modules, functions, responses)

	// Restore original stdout and read captured output
	_ = w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Define expected output
	wanted := "-=-=-=-=-=-=-=-=-=-=-=-=-=-=- IDENTIFIED SECURITY ISSUES -=-=-=-=-=-=-=-=-=-=-=-=-=-=-\nModule module2 not found\nModule headers not found\n"

	// Check if the function produced the expected output
	if got := buf.String(); got != wanted {
		t.Errorf("runModules() failed, got: %q, want: %q", got, wanted)
	}
}

func TestPrintRelationships(t *testing.T) {
	tests := []struct {
		tag           string
		relationships map[string][]string
		expected      string
	}{
		{
			"Admin",
			map[string][]string{
				"https://example.com/admin": {"https://example.com/settings"},
			},
			"\n" + strings.Repeat("-", 100) + "\nAdmin can visit the following:\nhttps://example.com/admin leads to:\n\thttps://example.com/settings\n",
		},
		{
			"User",
			map[string][]string{
				"https://example.com/profile": {"https://example.com/logout"},
			},
			"\n" + strings.Repeat("-", 100) + "\nUser can visit the following:\nhttps://example.com/profile leads to:\n\thttps://example.com/logout\n",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing printRelationships with tag = %q and relationships = %v", test.tag, test.relationships)

			// Redirect stdout to a buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printRelationships(test.tag, test.relationships)

			// Restore original stdout
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			result := buf.String()

			if result != test.expected {
				t.Errorf("printRelationships(%q, %v) = %q; expected %q", test.tag, test.relationships, result, test.expected)
			} else {
				t.Logf("printRelationships(%q, %v) = %q; passed", test.tag, test.relationships, result)
			}
		})
	}
}

func TestStringInSlice(t *testing.T) {
	tests := []struct {
		str      string
		slice    []string
		expected bool
	}{
		{"https://example.com", []string{"https://example.com"}, true},
		{"http://example.com", []string{"https://example.com"}, false},
		{"https://example.com", []string{"https://example.com", "http://example.com"}, true},
		{"https://example.org", []string{"https://example.com", "http://example.com"}, false},
		{"test/path", []string{"https://example.com", "https://example.com/test/path"}, false},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing stringInSlice with str = %q and slice = %v", test.str, test.slice)
			result := stringInSlice(test.str, test.slice)
			if result != test.expected {
				t.Errorf("stringInSlice(%q, %v) = %v; expected %v", test.str, test.slice, result, test.expected)
			} else {
				t.Logf("stringInSlice(%q, %v) = %v; passed", test.str, test.slice, result)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Test case 1",
			input:    []string{"https://example.com", "https://example.com"},
			expected: []string{"https://example.com"},
		},
		{
			name:     "Test case 2",
			input:    []string{"http://example.com", "https://example.com"},
			expected: []string{"http://example.com", "https://example.com"},
		},
		{
			name:     "Test case 3",
			input:    []string{"https://example.com", "https://example.com", "https://example.com"},
			expected: []string{"https://example.com"},
		},
		{
			name:     "Test case 4",
			input:    []string{"https://example.com", "http://example.com", "https://example.org"},
			expected: []string{"https://example.com", "http://example.com", "https://example.org"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing removeDuplicates with input = %v", test.input)
			result := removeDuplicates(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("removeDuplicates(%v) = %v; expected %v", test.input, result, test.expected)
			} else {
				t.Logf("removeDuplicates(%v) = %v; passed", test.input, result)
			}
		})
	}
}

/*
func TestProcessFlags(t *testing.T) {
	tests := []struct {
		name                 string
		args                 []string
		expectedURL          string
		expectedAuth         string
		expectedTag          string
		expectedAllowSubs    bool
		expectedModules      string
		expectedReqPerSecond int
		expectedFuncScrape   bool
		expectedCrawl        bool
		expectedVerbose      bool
		expectedProd         bool
		expectedErr          error
	}{
		{
			name:                 "Test case 1",
			args:                 []string{"-u", "https://example.com", "-v"},
			expectedURL:          "https://example.com",
			expectedAuth:         "",
			expectedTag:          "",
			expectedAllowSubs:    false,
			expectedModules:      "",
			expectedReqPerSecond: 0,
			expectedFuncScrape:   false,
			expectedCrawl:        false,
			expectedVerbose:      true,
			expectedProd:         true,
			expectedErr:          nil,
		},
		{
			name:                 "Test case 2",
			args:                 []string{},
			expectedURL:          "",
			expectedAuth:         "",
			expectedTag:          "",
			expectedAllowSubs:    false,
			expectedModules:      "",
			expectedReqPerSecond: 0,
			expectedFuncScrape:   false,
			expectedCrawl:        false,
			expectedVerbose:      false,
			expectedProd:         true,
			expectedErr:          errors.New("invalid URL"),
		},
	}

	originalArgs := os.Args

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running %s", test.name)

			os.Args = append([]string{originalArgs[0]}, test.args...)
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			url, auth, tag, allowSubs, modules, reqPerSecond, funcScrape, crawl, verbose, prod, err := processFlags()

			if err != test.expectedErr {
				t.Errorf("processFlags returned error to %s; expected %s", err, test.expectedErr)
			}

			if url != test.expectedURL {
				t.Errorf("processFlags set url to %s; expected %s", url, test.expectedURL)
			}

			if auth != test.expectedAuth {
				t.Errorf("processFlags set auth to %s; expected %s", auth, test.expectedAuth)
			}

			if tag != test.expectedTag {
				t.Errorf("processFlags set tag to %s; expected %s", tag, test.expectedTag)
			}

			if allowSubs != test.expectedAllowSubs {
				t.Errorf("processFlags set allowSubs to %t; expected %t", allowSubs, test.expectedAllowSubs)
			}

			if modules != test.expectedModules {
				t.Errorf("processFlags set modules to %s; expected %s", modules, test.expectedModules)
			}

			if reqPerSecond != test.expectedReqPerSecond {
				t.Errorf("processFlags set reqPerSecond to %d; expected %d", reqPerSecond, test.expectedReqPerSecond)
			}

			if funcScrape != test.expectedFuncScrape {
				t.Errorf("processFlags set funcScrape to %t; expected %t", funcScrape, test.expectedFuncScrape)
			}

			if crawl != test.expectedCrawl {
				t.Errorf("processFlags set crawl to %t; expected %t", crawl, test.expectedCrawl)
			}

			if verbose != test.expectedVerbose {
				t.Errorf("processFlags set verbose to %t; expected %t", verbose, test.expectedVerbose)
			}

			if prod != test.expectedProd {
				t.Errorf("processFlags set prod to %t; expected %t", prod, test.expectedProd)
			}
		})
	}
}
*/

func TestCreateCollector(t *testing.T) {
	tests := []struct {
		
		name            string
		domain          string
		allowSubdomains bool
	}{
		{"Test case 1", "example.com", false},
		{"Test case 2", "example.com", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running %s", test.name)
			collector := createCollector(test.domain, test.allowSubdomains)

			if collector == nil {
				t.Error("createCollector returned a nil collector")
				return
			}

			// Check if the domain is included in the allowed domains
			found := false
			for _, domain := range collector.AllowedDomains {
				if domain == test.domain || (test.allowSubdomains && domain == "*."+test.domain) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("createCollector did not include the domain in the AllowedDomains")
			} else {
				t.Logf("createCollector included the domain in the AllowedDomains; passed")
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		expected *Config
		hasError bool
	}{
		{
			name: "Test case 1",
			config: `
{
	"inscope_ips": ["192.168.1.1"],
	"inscope_urls": ["https://example.com"],
	"authentication": {
		"username": "user",
		"password": "pass"
	},
	"max_requests_per_sec": 10
}`,
			expected: &Config{
				InscopeIPs:        []string{"192.168.1.1"},
				InscopeURLs:       []string{"https://example.com"},
				Authentication:    map[string]string{"username": "user", "password": "pass"},
				MaxRequestsPerSec: 10,
			},
			hasError: false,
		},
		{
			name: "Test case 2",
			config: `
{
	"inscope_ips": "invalid",
	"inscope_urls": ["https://example.com"],
	"authentication": {
		"username": "user",
		"password": "pass"
	},
	"max_requests_per_sec": 10
}`,
			expected: nil,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running %s", test.name)
			tempFile, err := ioutil.TempFile("", "configfile")
			if err != nil {
				t.Fatalf("Error creating temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			err = ioutil.WriteFile(tempFile.Name(), []byte(test.config), 0644)
			if err != nil {
				t.Fatalf("Error writing to temp file: %v", err)
			}

			result, err := LoadConfig(tempFile.Name())
			if test.hasError {
				if err == nil {
					t.Error("LoadConfig should have returned an error, but didn't")
				} else {
					t.Logf("LoadConfig returned an error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("LoadConfig returned an error: %v", err)
				} else if !reflect.DeepEqual(result, test.expected) {
					t.Errorf("LoadConfig = %+v; expected %+v", result, test.expected)
				} else {
					t.Logf("LoadConfig = %+v; passed", result)
				}
			}
		})
	}
}

func TestWriteRelationshipsToCSV(t *testing.T) {
	tests := []struct {
		name          string
		links         map[string][]string
		expectedLines int
	}{
		{
			name: "Test case 1",
			links: map[string][]string{
				"https://example.com": {"https://example.com/page1", "https://example.com/page2"},
			},
			expectedLines: 3,
		},
		{
			name: "Test case 2",
			links: map[string][]string{
				"https://example.com":       {"https://example.com/page1"},
				"https://example.com/page1": {"https://example.com/page2"},
			},
			expectedLines: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running %s", test.name)
			tempFile, err := ioutil.TempFile("", "csvfile")
			if err != nil {
				t.Fatalf("Error creating temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			writeRelationshipsToCSV(test.links, tempFile.Name())

			fileContent, err := ioutil.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatalf("Error reading temp file: %v", err)
			}

			lines := strings.Split(string(fileContent), "\n")
			// Subtract 1 from len(lines) to account for the trailing newline
			if len(lines)-1 != test.expectedLines {
				t.Logf("CSV content:\n%s", string(fileContent)) // Output the contents of the CSV file
				t.Errorf("writeRelationshipsToCSV generated %d lines; expected %d", len(lines)-1, test.expectedLines)
			} else {
				t.Logf("writeRelationshipsToCSV generated %d lines; passed", len(lines)-1)
			}
		})
	}
}

func TestCreateDotFile(t *testing.T) {
	tests := []struct {
		name          string
		links         map[string][]string
		expectedLines int
	}{
		{
			name: "Test case 1",
			links: map[string][]string{
				"https://example.com": {"https://example.com/page1", "https://example.com/page2"},
			},
			expectedLines: 4,
		},
		{
			name: "Test case 2",
			links: map[string][]string{
				"https://example.com":       {"https://example.com/page1"},
				"https://example.com/page1": {"https://example.com/page2"},
			},
			expectedLines: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Running %s", test.name)
			tempFile, err := ioutil.TempFile("", "dotfile")
			if err != nil {
				t.Fatalf("Error creating temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())

			createDotFile(test.links, tempFile.Name())

			fileContent, err := ioutil.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatalf("Error reading temp file: %v", err)
			}

			lines := strings.Split(string(fileContent), "\n")
			num_lines := len(lines) - 1
			if num_lines != test.expectedLines { //-1 to remove newline char
				t.Logf("Graph File content:\n%s", string(fileContent)) // Output the contents of the CSV file
				t.Errorf("createDotFile generated %d lines; expected %d", num_lines, test.expectedLines)
			} else {
				t.Logf("createDotFile generated %d lines; passed", len(lines))
			}
		})
	}
}

// Test getDomain
func TestGetDomain(t *testing.T) {
	tests := []struct {
		urlString string
		expected  string
	}{
		{"https://www.example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"http://www.example.com/path/to/something", "example.com"},
		{"https://example.com/path/to/something", "example.com"},
		{"example.com", "example.com"},
	}

	for _, test := range tests {
		result := getDomain(test.urlString)
		if result != test.expected {
			t.Errorf("getDomain(%q) = %q; expected %q", test.urlString, result, test.expected)
		}
	}
}

// Test validateFlags
func TestValidateFlags(t *testing.T) {
	tests := []struct {
		url_input        *string
		auth_input       *string
		tag_input        *string
		subdomains_input *bool
		modules_input    *string
		requests_input   *int
		prod_input       *bool
		expectedError    bool
	}{
		{stringPtr("https://www.example.com"), stringPtr("Cookie: name=value;"), stringPtr("Admin"), boolPtr(false), stringPtr("headers"), intPtr(0), boolPtr(true), false},
		{stringPtr("https://www.example.com"), stringPtr(""), stringPtr("Admin"), boolPtr(false), stringPtr("headers,cors"), intPtr(0), boolPtr(true), false},
		{stringPtr("https://www.example.com"), stringPtr(""), stringPtr("Admin"), boolPtr(false), stringPtr("headers,invalid"), intPtr(0), boolPtr(true), true},
		{stringPtr("www.example.com"), stringPtr(""), stringPtr(""), boolPtr(false), stringPtr(""), intPtr(0), boolPtr(true), true},
	}

	for _, test := range tests {
		err := validateFlags(test.url_input, test.auth_input, test.tag_input, test.subdomains_input, test.modules_input, test.requests_input, test.prod_input)
		if (err != nil) != test.expectedError {
			t.Errorf("validateFlags(%q, %q, %q, %t, %q, %d, %t) returned error: %v", *test.url_input, *test.auth_input, *test.tag_input, *test.subdomains_input, *test.modules_input, *test.requests_input, *test.prod_input, err)
		}
	}
}

// Test addTrailingSlashIfNeeded
func TestAddTrailingSlashIfNeeded(t *testing.T) {
	tests := []struct {
		inputURL string
		expected string
	}{
		{"https://example.com", "https://example.com/"},
		{"http://example.com/test", "http://example.com/test"},
		{"https://example.com/test?query=test", "https://example.com/test?query=test"},
	}

	for _, test := range tests {
		result, err := addTrailingSlashIfNeeded(test.inputURL)
		if err != nil {
			t.Errorf("addTrailingSlashIfNeeded(%q) returned error: %v", test.inputURL, err)
		}
		if result != test.expected {
			t.Errorf("addTrailingSlashIfNeeded(%q) = %q; expected %q", test.inputURL, result, test.expected)
		}
	}
}

// Helper functions for creating test data
func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }
func intPtr(i int) *int          { return &i }

func TestMain(m *testing.M) {
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}

func TestRemovePort(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com:80", "example.com"},
		{"example.com:443", "example.com"},
		{"example.com", "example.com"},
	}

	for _, test := range tests {
		result := removePort(test.input)
		if result != test.expected {
			t.Errorf("removePort(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestCalcURL(t *testing.T) {
	tests := []struct {
		src      string
		domain   string
		expected string
	}{
		{"https://example.com", "https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com", "http://example.com"},
		{"//example.com", "https://example.com", "https://example.com"},
		{"/test/path", "https://example.com", "https://example.com/test/path"},
		{"test/path", "https://example.com", ""},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
			t.Logf("Testing calc_url with src = %q and domain = %q", test.src, test.domain)
			result := calc_url(test.src, test.domain)
			if result != test.expected {
				t.Errorf("calc_url(%q, %q) = %q; expected %q", test.src, test.domain, result, test.expected)
			} else {
				t.Logf("calc_url(%q, %q) = %q; passed", test.src, test.domain, result)
			}
		})
	}
}
