package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

var (
	client  *resty.Client
	links   []string
	baseUrl string
)

type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc string `xml:"loc"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <filename> <extra endings>")
		return
	}

	fmt.Println("Welcome to the scraping")
	baseUrl = os.Args[1]
	getHttp("/")

	if len(os.Args) > 2 {
		for _, value := range os.Args[2:] {
			getHttp(value)
		}
	}

	generateSitemap()
}

func Contains[T comparable](arr []T, item T) bool {
	arrLen := len(arr)
	for i := 0; i < arrLen; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func getHttp(ending string) {
	if Contains(links, ending) {
		return
	}
	links = append(links, ending)
	client = resty.New()
	site, err := client.R().Get("https://" + baseUrl + ending)
	if err != nil {
		fmt.Println("get request error: ", err)
		return
	}
	respStr := site.String()

	findAnchors(respStr)
	stringToHtml(respStr, ending)
}

func findAnchors(respStr string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(respStr))
	if err != nil {
		fmt.Println("error parsing document: ", err)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			getHttp(href)
		} else {
			fmt.Println("No links found")
		}
	})
}

func dirPaths(path string) string {
	parts := strings.Split(path, "/")

	return strings.Join(parts[:len(parts)-1], "/")
}

func stringToHtml(htmlStr string, path string) {
	// neededDirs := dirPaths(baseUrl + path)
	err := os.MkdirAll(baseUrl+path, 0755)
	if err != nil {
		fmt.Println("error making directories: ")
		return
	}
	newFilePath := baseUrl + path + "/index.html"
	if path == "/" {
		newFilePath = baseUrl + "/index.html"
	}

	file, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println("os create error: ", err)
		return
	}

	defer file.Close()

	_, err = file.WriteString(htmlStr)
	if err != nil {
		fmt.Println("error writing to file: ", err)
		return
	}

	fmt.Println("HTML content saved to ", newFilePath)
}

func generateSitemap() {
	sitemap := Sitemap{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  make([]URL, 0),
	}

	for _, link := range links {
		sitemap.URLs = append(sitemap.URLs, URL{Loc: "https://" + baseUrl + link})
	}

	output, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		fmt.Println("error marshaling sitemap: ", err)
		return
	}

	file, err := os.Create(baseUrl + "/sitemap.xml")
	if err != nil {
		fmt.Println("error creating sitemap file: ", err)
		return
	}
	defer file.Close()

	file.WriteString(xml.Header)
	file.Write(output)

	fmt.Println("Sitemap generated and saved to ", baseUrl+"/sitemap.xml")
}
