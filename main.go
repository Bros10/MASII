package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gocolly/colly/v2"
	"github.com/olekukonko/tablewriter"
)

//Multiple global structs to store all HTTP responses, Page Data and found security issues.

type responseData struct {
	URL      string
	Response string
	Headers  map[string]string
}

type PageData struct {
	URL         string
	elements    string
	AccessLevel []string
}

type SecurityIssues struct {
	URL            string
	missingHeaders []string
}

//This struct is modifable to allow for new settings within ```config.json```

type Config struct {
	InscopeIPs        []string          `json:"inscope_ips"`
	InscopeURLs       []string          `json:"inscope_urls"`
	Authentication    map[string]string `json:"authentication"`
	MaxRequestsPerSec int               `json:"max_requests_per_sec"`
}

var delayBetweenVisits time.Duration

func main() {
	pageDataMap := make(map[string]PageData)

	// Create a new table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Page URL", "HTML Elements Present", "Access Level"})
	// Set column widths explicitly
	table.SetAutoWrapText(false)

	var responses []responseData
	functions := map[string]func(responses []responseData, verbose bool){ // Define a map to store pointers to functions
		"headers": headers,
		"cors":    cors,
	}

	urlString, auth, tags, allow_subdomains, modules, requests_per_second, func_scrape, crawl, verbose, prod, err := processFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	prod = prod

	domain := getDomain(urlString)
	domain = domain
	delayBetweenVisits = time.Duration(float64(1) / float64(requests_per_second) * float64(time.Second))

	if auth != "" {
		//If auth is passed.
		//Do unauth scan and then do auth scans

		parts := strings.SplitN(auth, ":", 2)
		auth_type := parts[0]
		auth = parts[1]
		tag := strings.Split(tags, ",")
		tag = append(tag, "Unauth")
		cookies := strings.Split(auth, ",")
		cookies = append(cookies, "")
		num_roles := 0
		count := 0
		for _, cookie := range cookies {
			//So this loops through Cookies which I have to loop through the tags.
			exact_tag := tag[count]
			fmt.Println(exact_tag)

			count = count + 1

			c := createCollector(domain, allow_subdomains)
			c.Limit(&colly.LimitRule{
				Parallelism: 1,
				Delay:       100 * time.Second,
			})

			//Now we call the entire crawling aspect
			// Keep track of visited pages using a map
			visited := make(map[string]bool)
			// Keep track of the relationships between pages using a map
			relationships := make(map[string][]string)

			ticker := time.NewTicker(time.Duration(1e6/requests_per_second) * time.Microsecond)
			//fmt.Println(ticker)
			ticker = ticker

			setCookieCallback(c, parts, auth_type, cookie)
			c.OnRequest(func(r *colly.Request) {
				//Sets the cookie for every request.
				r.Headers.Set(auth_type, cookie)
			})
			// Set a callback function for the OnResponse event, which is called
			// whenever the crawler receives a response from a server
			c.OnResponse(func(r *colly.Response) {
				headers := make(map[string]string)
				for key, values := range *r.Headers {
					headers[key] = values[0]
				}

				outputURL, err := addTrailingSlashIfNeeded(r.Request.URL.String())
				if err != nil {
					fmt.Println("Error loading URL:", err)
					return
				}

				responses = append(responses, responseData{
					URL:      outputURL, //Need to ensure that "/" is appended on it. So we don't get google.com and google.com/
					Response: string(r.Body),
					Headers:  headers,
				})

				visitJavascriptEndpoints(c, r, urlString, responses, relationships)
				// Set a callback function for the OnHTML event, which is called
				// whenever the crawler encounters a page with HTML content
				c.OnHTML("html", func(e *colly.HTMLElement) {
					visitPage(c, r, e, visited, relationships, func_scrape, table, exact_tag, pageDataMap)
				})
			})

			// Start the crawl by visiting the specified URL
			c.Visit(urlString)
			time.Sleep(delayBetweenVisits)

			createDotFile(relationships, tag[num_roles]+"graph.dot")

			err = writeRelationshipsToCSV(relationships, tag[num_roles]+"relationships.csv")
			if err != nil {
				fmt.Println("Error writing CSV:", err)
			}

			config, err := LoadConfig("config.json")
			if err != nil {
				fmt.Println("Error loading config file:", err)
				return
			}

			config = config

			if crawl == true {
				//printRelationships(tag[num_roles], relationships)
			}

			fmt.Println("")

			if modules != "" {
				runModules(verbose, modules, functions, responses)
			}

			//If run modules arg provided
			num_roles = num_roles + 1
		}

	} else {
		tag := "Unauth"

		c := createCollector(domain, allow_subdomains)

		//Now we call the entire crawling aspect

		// Keep track of visited pages using a map
		visited := make(map[string]bool)
		// Keep track of the relationships between pages using a map
		relationships := make(map[string][]string)

		var parts []string
		auth_type := ""
		cookie := ""
		setCookieCallback(c, parts, auth_type, cookie)

		// Set a callback function for the OnResponse event, which is called
		// whenever the crawler receives a response from a server
		c.OnResponse(func(r *colly.Response) {
			headers := make(map[string]string)
			for key, values := range *r.Headers {
				headers[key] = values[0]
			}

			outputURL, err := addTrailingSlashIfNeeded(r.Request.URL.String())
			if err != nil {
				fmt.Println("Error loading URL:", err)
				return
			}

			responses = append(responses, responseData{
				URL:      outputURL, //Need to ensure that "/" is appended on it. So we don't get google.com and google.com/
				Response: string(r.Body),
				Headers:  headers,
			})

			visitJavascriptEndpoints(c, r, urlString, responses, relationships)
			// Set a callback function for the OnHTML event, which is called
			// whenever the crawler encounters a page with HTML content
			c.OnHTML("html", func(e *colly.HTMLElement) {
				visitPage(c, r, e, visited, relationships, func_scrape, table, tag, pageDataMap)
			})
		})

		// Start the crawl by visiting the specified URL
		println("Visting", urlString)
		c.Visit(urlString)
		time.Sleep(delayBetweenVisits)

		createDotFile(relationships, "unauth"+"graph.dot")

		err = writeRelationshipsToCSV(relationships, "unauth"+"relationships.csv")
		if err != nil {
			fmt.Println("Error writing CSV:", err)
		}

		config, err := LoadConfig("config.json")
		if err != nil {
			fmt.Println("Error loading config file:", err)
			return
		}

		config = config

		if crawl == true {
			printRelationships("Unauth", relationships)
		}

		fmt.Println("")

		if modules != "" {
			runModules(verbose, modules, functions, responses)
		}

	}

	// Iterate through the pageDataMap and output the data
	for _, pageData := range pageDataMap {
		table.Append([]string{pageData.URL, pageData.elements, strings.Join(pageData.AccessLevel, ", ")})
	}

	table.SetAutoWrapText(false)
	//Render the table
	table.Render()

}

func runModules(verbose bool, modules string, functions map[string]func([]responseData, bool), responses []responseData) {
	fmt.Println("-=-=-=-=-=-=-=-=-=-=-=-=-=-=- IDENTIFIED SECURITY ISSUES -=-=-=-=-=-=-=-=-=-=-=-=-=-=-")
	//start := time.Now()
	if modules == "all" {
		//Default value, meaning run all modules
		for _, f := range functions {
			go f(responses, verbose) // This will now run concurrently
		}
	} else {
		modules := strings.Split(modules, ",")
		// Iterate through the modules slice
		for _, module := range modules {
			// Check if the module exists in the functions map
			if f, ok := functions[module]; ok {
				go f(responses, verbose) // This will now run concurrently
			} else {
				fmt.Printf("Module %s not found\n", module)
			}
		}
	}
	//fmt.Printf("Concurrency Total time: %v\n", time.Since(start))
}

func printRelationships(tag string, relationships map[string][]string) {
	fmt.Println("")
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println(tag, "can visit the following:")
	for k, v := range relationships {
		fmt.Printf("%s leads to:\n", k)
		for _, val := range v {
			fmt.Printf("\t%s\n", val)
		}
	}
}

func visitPage(c *colly.Collector, r *colly.Response, e *colly.HTMLElement, visited map[string]bool, relationships map[string][]string, func_scrape bool, table *tablewriter.Table, tag string, pageDataMap map[string]PageData) {
	//Need to define access levels

	urlString := e.Request.URL.String()

	outputURL, err := addTrailingSlashIfNeeded(urlString)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		//fmt.Println("Input:", urlString, "Output:", outputURL)
	}

	urlString = outputURL

	// Check if the page has already been visited
	if visited[urlString] {
		// If the page has already been visited, skip it
		return
	}

	// If the page has not been visited, mark it as visited
	visited[urlString] = true

	// Visit all links on the page
	e.ForEach("a[href]", func(i int, element *colly.HTMLElement) {
		// Get the value of the href attribute (i.e. the link URL)
		link := element.Attr("href")

		// Visit the link if it is not already in the visited map
		if !visited[link] {
			c.Visit(e.Request.AbsoluteURL(link))
			time.Sleep(delayBetweenVisits)

			// Add the relationship between the current page and the linked page

			if urlString != e.Request.AbsoluteURL(link) {
				relationships[urlString] = append(relationships[urlString], e.Request.AbsoluteURL(link))
			} else {
				//fmt.Println("Equal")
			}
		}
	})

	// Initialize variables to store HTML elements for each URL
	var formElements, inputElements, buttonElements []string

	// Map out the forms, inputs, and buttons on the page
	e.ForEach("form", func(i int, element *colly.HTMLElement) {
		formElements = append(formElements, "form")
	})
	e.ForEach("input", func(i int, element *colly.HTMLElement) {
		inputElements = append(inputElements, "input")
	})
	e.ForEach("button", func(i int, element *colly.HTMLElement) {
		buttonElements = append(buttonElements, "button")
	})
	e.ForEach("textarea", func(i int, element *colly.HTMLElement) {
		buttonElements = append(buttonElements, "textarea")
	})
	e.ForEach("select", func(i int, element *colly.HTMLElement) {
		buttonElements = append(buttonElements, "select")
	})
	e.ForEach("datalist", func(i int, element *colly.HTMLElement) {
		buttonElements = append(buttonElements, "datalist")
	})
	e.ForEach("label", func(i int, element *colly.HTMLElement) {
		buttonElements = append(buttonElements, "label")
	})

	// Combine all the HTML elements for the current URL
	allElements := append(formElements, inputElements...)
	allElements = append(allElements, buttonElements...)

	// Remove duplicates
	uniqueElements := removeDuplicates(allElements)

	// Convert the unique HTML elements slice into a comma-separated string
	elementsString := strings.Join(uniqueElements, ", ")
	accessLevels := tag

	//Insert data into PageData here
	// Create a map to store instances of PageData with the URL as the key

	// Check if the URL already exists in the map
	if page, exists := pageDataMap[urlString]; exists {
		// If the URL exists, append the new access level to the AccessLevel slice
		page.AccessLevel = append(page.AccessLevel, accessLevels)
		pageDataMap[urlString] = page
	} else {
		// If the URL doesn't exist, create a new PageData instance and add it to the map
		pageDataMap[urlString] = PageData{
			URL:         urlString,
			elements:    elementsString,
			AccessLevel: []string{accessLevels},
		}
	}

	if func_scrape == true {
		//table.Render()
	}

	e.ForEach("script", func(i int, element *colly.HTMLElement) {
		src := element.Attr("src")

		// Visit the link if it is not already in the visited map
		if !visited[src] {
			c.Visit(e.Request.AbsoluteURL(src))
			time.Sleep(delayBetweenVisits)

			// Add the relationship between the current page and the linked page
			relationships[urlString] = append(relationships[urlString], e.Request.AbsoluteURL(src))
		}
	})
}

func stringInSlice(value string, slice []string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func visitJavascriptEndpoints(c *colly.Collector, r *colly.Response, urlString string, responses []responseData, relationships map[string][]string) {

	// Check if the response is a JavaScript file
	if strings.Contains(r.Headers.Get("Content-Type"), "application/javascript") {
		// Get the contents of the JavaScript file as a string
		js := string(r.Body)

		re := regexp.MustCompile(`"[^"]+"`)
		pos_endpoint := re.FindAllString(js, -1)
		for _, endpoint := range pos_endpoint {
			endpoint = endpoint[1 : len(endpoint)-1]
			first_char := endpoint[0:1]

			if strings.HasPrefix(endpoint, "https://") {
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "http://") {
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "//") {
				endpoint = "https:" + endpoint
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "/") {
				endpoint = urlString + endpoint
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else {
				for _, m := range first_char {
					if !unicode.IsLetter(m) {
						endpoint = urlString + "/" + endpoint
						c.Visit(endpoint)
						time.Sleep(delayBetweenVisits)

						relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

					}
				}
			}
		}
	}
}

// Remove duplicates function
func removeDuplicates(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, element := range elements {
		if !encountered[element] {
			encountered[element] = true
			result = append(result, element)
		}
	}

	return result
}

func handleResponse(r *colly.Response, c *colly.Collector, urlString string, responses []responseData, relationships map[string][]string) {
	headers := make(map[string]string)
	for key, values := range *r.Headers {
		headers[key] = values[0]
	}

	responses = append(responses, responseData{
		URL:      r.Request.URL.String(),
		Response: string(r.Body),
		Headers:  headers,
	})

	// Check if the response is a JavaScript file
	if strings.Contains(r.Headers.Get("Content-Type"), "application/javascript") {
		// Get the contents of the JavaScript file as a string
		js := string(r.Body)

		re := regexp.MustCompile(`"[^"]+"`)
		pos_endpoint := re.FindAllString(js, -1)
		for _, endpoint := range pos_endpoint {
			//endpoint = endpoint[len(endpoint)-1:]
			endpoint = endpoint[1 : len(endpoint)-1]
			first_char := endpoint[0:1]

			if strings.HasPrefix(endpoint, "https://") {
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "http://") {
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "//") {
				endpoint = "https:" + endpoint
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				// Add the relationship between the current page and the endpoint
				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else if strings.HasPrefix(endpoint, "/") {
				endpoint = urlString + endpoint
				c.Visit(endpoint)
				time.Sleep(delayBetweenVisits)

				relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

			} else {
				for _, m := range first_char {
					if !unicode.IsLetter(m) {
						endpoint = urlString + "/" + endpoint
						c.Visit(endpoint)
						time.Sleep(delayBetweenVisits)

						relationships[r.Request.URL.String()] = append(relationships[r.Request.URL.String()], endpoint)

					} else {
						//fmt.Println("Does not begin with letters")
					}
				}
			}
		}
	}
}

func setCookieCallback(c *colly.Collector, parts []string, auth_type string, cookie string) {
	c.OnRequest(func(r *colly.Request) {
		if auth_type != "" && cookie != "" {
			r.Headers.Set(auth_type, cookie)
		}
	})
}

func processFlags() (string, string, string, bool, string, int, bool, bool, bool, bool, error) {
	url_input := flag.String(("u"), "", "Specify the target URL for the tool to assess (eg., https://example.com)")
	auth_input := flag.String(("a"), "", "Provide authentication headers or cookies (e.g., Cookie: name1=value_1;)")
	tag_input := flag.String(("t"), "", "Label the Cookie or Auth token with a user role (e.g., Admin or Customer)")
	subdomains_input := flag.Bool(("s"), false, "Enable the tool to crawl all subdomains when the flag is provided")
	modules_input := flag.String(("m"), "", "Select the security modules to be executed on the website (e.g., headers, cors, or all)")
	requests_input := flag.Int("r", 0, "Specify the maximum number of requests sent per second by the tool")
	func_input := flag.Bool(("f"), false, "Enable functionality scraping feature")
	crawl_input := flag.Bool(("c"), false, "Enable crawling for all provided user roles")
	verbose_input := flag.Bool(("v"), false, "Display verbose output, including all HTTP responses related to security issues")
	prod_input := flag.Bool(("prod"), true, "Indicate whether the website is in a production environment")
	flag.Parse()

	err := validateFlags(url_input, auth_input, tag_input, subdomains_input, modules_input, requests_input, prod_input)
	if err != nil {
		return "", "", "", false, "", 0, false, false, false, false, err
	}

	urlString := *url_input
	auth := *auth_input
	tag := *tag_input
	allow_subdomains := *subdomains_input
	modules := *modules_input
	requests_per_second := *requests_input
	prod := *prod_input
	func_scrape := *func_input
	crawl := *crawl_input
	verbose := *verbose_input

	return urlString, auth, tag, allow_subdomains, modules, requests_per_second, func_scrape, crawl, verbose, prod, nil
}

func validateFlags(url_input *string, auth_input *string, tag_input *string, subdomains_input *bool, modules_input *string, requests_input *int, prod_input *bool) error {
	// validate URL
	if !strings.HasPrefix(*url_input, "http://") && !strings.HasPrefix(*url_input, "https://") {
		return fmt.Errorf("invalid URL: %s", *url_input)
	}

	// validate auth_input format
	if *auth_input != "" && !strings.Contains(*auth_input, "Cookie:") && !strings.Contains(*auth_input, "Authorization:") {
		return fmt.Errorf("invalid auth_input format: %s. Must be either a Cookie or an Authorization token.", *auth_input)
	}

	if *auth_input != "" {
		tags := strings.Split(*tag_input, ",")
		for _, tag := range tags {
			if tag != "tag" && !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(tag) {
				return fmt.Errorf("invalid tag_input format: %s. Must be 'tag' or a string of alphanumeric characters and underscores", tag)
			}
		}
	}

	// validate modules_input
	modules := strings.Split(*modules_input, ",")
	for _, module := range modules {
		if module != "headers" && module != "cors" && module != "" {
			return fmt.Errorf("invalid module: %s. Must be 'headers' or 'cors'", module)
		}
	}

	// validate requests_input
	if *requests_input < 0 {
		return fmt.Errorf("invalid requests_input: %d. Must be non-negative", *requests_input)
	}

	return nil
}

func getDomain(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		log.Fatal(err)
	}
	var domain string
	if u.Host != "" {
		domain = strings.TrimPrefix(u.Host, "www.")
	} else {
		// if Host is empty, check the input string
		if strings.HasPrefix(urlString, "http") {
			// if the input string starts with "http" it is assumed that the domain is present after the 3rd '/'
			domain = strings.Split(urlString, "/")[2]

		} else {
			// if the input string doesn't start with "http" it is assumed that the domain is present at the beginning of the string
			domain = strings.Split(urlString, "/")[0]
		}
		domain = strings.TrimPrefix(domain, "www.")
	}
	domain = removePort(domain)
	return domain
}

// function to create a collector with allowed domains based on user's choice
func createCollector(domain string, allowSubdomains bool) *colly.Collector {
	c := colly.NewCollector()

	if allowSubdomains == true {
		c = colly.NewCollector(
			colly.AllowedDomains(domain, "*."+domain),
		)
	} else {
		c = colly.NewCollector(
			colly.AllowedDomains(domain),
		)
	}
	c.Limit(&colly.LimitRule{
		Delay: 10 * time.Second,
	})

	return c
}

func headers(responses []responseData, verbose bool) {
	SecurityIssuesMap := make(map[string]SecurityIssues)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"URL", "Missing Security Headers"})

	securityHeaders := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Content-Security-Policy",
		"Referrer-Policy",
		"Permissions-Policy",
	}
	for _, d := range responses {
		var missingHeaders []string
		missingSecurityHeaders := false

		if !strings.Contains(d.Headers["Content-Type"], "text/html") {
			continue
		}

		for _, header := range securityHeaders {
			if _, ok := d.Headers[header]; !ok {
				missingSecurityHeaders = true
				missingHeaders = append(missingHeaders, header)
			}
		}

		// Output the URL only if it has "text/html" Content-Type
		// and is missing one or more security headers
		if missingSecurityHeaders {
			//fmt.Println(d.URL, "is missing the following security Headers:")
			//for _, header := range missingHeaders {
			//fmt.Printf("  - %s\n", header)
			if issue, exists := SecurityIssuesMap[d.URL]; exists {
				// If the URL exists, append the new access level to the AccessLevel slice
				for _, header := range missingHeaders {
					issue.missingHeaders = append(issue.missingHeaders, header)
				}
				SecurityIssuesMap[d.URL] = issue
			} else {
				for _, header := range missingHeaders {
					// If the URL doesn't exist, create a new PageData instance and add it to the map
					SecurityIssuesMap[d.URL] = SecurityIssues{
						URL:            d.URL,
						missingHeaders: []string{header},
					}
				}

			}
		}
		if verbose {
			fmt.Println("HTTP response headers:")
			for header, value := range d.Headers {
				fmt.Printf("  %s: %s\n", header, value)
			}
		}
	}

	// Iterate through the pageDataMap and output the data
	for _, pageData := range SecurityIssuesMap {
		table.Append([]string{pageData.URL, strings.Join(removeDuplicates(pageData.missingHeaders), ",")})
	}
	table.SetAutoWrapText(false)
	table.Render()
}

func cors(responses []responseData, verbose bool) {

	for _, d := range responses {
		if !strings.Contains(d.Headers["Content-Type"], "text/html") {
			continue
		}

		req, err := http.NewRequest("GET", d.URL, nil)
		if err != nil {
			fmt.Printf("Error creating request for %s: %s\n", d.URL, err)
			continue
		}

		req.Header.Set("Origin", "*")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request to %s: %s\n", d.URL, err)
			continue
		}

		if resp.Header.Get("Access-Control-Allow-Origin") == "*" &&
			resp.Header.Get("Access-Control-Allow-Credentials") == "true" {
			fmt.Printf("%s is vulnerable to CORS attack\n", d.URL)
			if verbose {
				fmt.Println("HTTP response headers:")
				for header, value := range d.Headers {
					fmt.Printf("  %s: %s\n", header, value)
				}
			}
		} else {
			fmt.Printf("%s is not vulnerable to CORS attack\n", d.URL)
		}

		resp.Body.Close()
	}
}

func struct_test(responses []responseData) {
	//fmt.Println("STRUCT TEST")
	for _, d := range responses {
		//fmt.Printf("URL: %s\nResponse: %s%s\n", d.URL, d.Headers, d.Response)
		d = d
	}
}

func addTrailingSlashIfNeeded(inputURL string) (string, error) {
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	if parsedURL.Path == "" {
		inputURL += "/"
	}

	return inputURL, nil
}

func removePort(url string) string {
	// Splitting the url by ":" and taking the first part of the split, if the split length is greater than 1
	if len(strings.Split(url, ":")) > 1 {
		url = strings.Split(url, ":")[0]
	}
	return url
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}

func writeRelationshipsToCSV(relationships map[string][]string, filename string) error {
	// Open the file for writing
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a new CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	err = writer.Write([]string{"Parent", "Child"})
	if err != nil {
		return err
	}

	// Write the data rows
	for parent, children := range relationships {
		for _, child := range children {
			err := writer.Write([]string{parent, child})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createDotFile(relationships map[string][]string, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	file.WriteString("digraph website {\n")
	for parent, children := range relationships {
		for _, child := range children {
			file.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", parent, child))
		}
	}
	file.WriteString("}\n")
}

func calc_url(src string, domain string) string {
	first_char := src[0:1]
	//fmt.Println(src)
	var url string
	if strings.HasPrefix(src, "https://") {
		url := src
		//fmt.Println(url)
		return url
		//8 char
	} else if strings.HasPrefix(src, "http://") {
		url := src
		//fmt.Println(url)
		return url
		//7 char
	} else if strings.HasPrefix(src, "//") {
		url := "https:" + src
		//fmt.Println(url)
		return url
		//2 char
	} else if strings.HasPrefix(src, "/") {
		url := domain + src
		//fmt.Println("URL:", url)
		return url
		//1 char
	} else {
		for _, r := range first_char {
			if !unicode.IsLetter(r) {
				// admin/login
				url := domain + "/" + src
				//fmt.Println(url)
				return url
			} else {
				//fmt.Println("Does not begin with letters")
				return ""
			}
		}
	}
	return url
}
