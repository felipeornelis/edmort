// package main

// import (
// 	"bufio"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"regexp"
// 	"strings"
// )

// var wishlist = "products.txt"

// type product struct {
// 	name  string
// 	price string
// }

// func Request(url string) (string, error) {
// 	req, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		return "", err
// 	}

// 	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

// 	client := &http.Client{}

// 	res, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}

// 	defer res.Body.Close()

// 	body, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	return bytes.NewBuffer(body).String(), nil
// }

// func Track(body string) product {
// 	priceRegex := regexp.MustCompile(`<span id="price"(.*?)>(.*?)</span>`)
// 	titleRegex := regexp.MustCompile(`<span id="productTitle"(.*?)>(.*?)</span>`)

// 	priceMatch := priceRegex.FindStringSubmatch(body)
// 	titleMatch := titleRegex.FindStringSubmatch(body)

// 	if len(priceMatch) > 1 && len(titleMatch) > 1 {
// 		return product{
// 			name:  strings.TrimSpace(titleMatch[2]),
// 			price: priceMatch[2],
// 		}
// 	}

// 	return product{}
// }

// func main() {
// 	products := make(map[string]product)

// 	f, err := os.Open(wishlist)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer f.Close()

// 	scanner := bufio.NewScanner(f)
// 	for scanner.Scan() {

// 		doc, err := Request(string(scanner.Text()))
// 		if err != nil {
// 			panic(err)
// 		}

// 		p := Track(doc)

// 		products[p.name] = p
// 	}

//		fmt.Printf("%v products tracked\n", len(products))
//		for _, item := range products {
//			fmt.Printf("%q costs %s\n", item.name, item.price)
//		}
//	}
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	bucketFile = "bucket.txt"
)

func main() {
	tracker := NewTracker(bucketFile)

	tracker.Run()
}

type tracker struct {
	bucket string
}

type item struct {
	name  string
	price int
}

func NewTracker(bucket string) *tracker {
	return &tracker{
		bucket: bucket,
	}
}

func (t tracker) bucketItemsURL() ([]string, error) {
	itemsURL := []string{}

	f, err := os.Open(t.bucket)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		itemsURL = append(itemsURL, scanner.Text())
	}

	return itemsURL, nil

}

func (t tracker) request(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (t tracker) parser(buf []byte) (item, error) {
	body := strings.TrimSpace(bytes.NewBuffer(buf).String())

	priceRegex := regexp.MustCompile(`<span id="price"(.*?)>(.*?)</span>`)
	titleRegex := regexp.MustCompile(`<span id="productTitle"(.*?)>(.*?)</span>`)

	priceMatch := priceRegex.FindStringSubmatch(body)
	titleMatch := titleRegex.FindStringSubmatch(body)

	if len(priceMatch) < 1 || len(titleMatch) < 1 {
		return item{}, errors.New("tracker.parser: error while trying parse body")
	}

	i := item{
		name:  strings.TrimSpace(titleMatch[2]),
		price: parsePrice(priceMatch[2]),
	}

	return i, nil
}

func parsePrice(price string) int {
	unformatedPriceString := strings.Split(price, " ")
	priceString := strings.Split(unformatedPriceString[1], ",")

	parsedRealToCentsInt, _ := strconv.Atoi(priceString[0])
	parsedCentsToInt, _ := strconv.Atoi(priceString[1])

	priceInCents := (parsedRealToCentsInt * 100) + parsedCentsToInt

	return priceInCents
}

func (t tracker) Run() {
	bucket, err := t.bucketItemsURL()
	if err != nil {
		panic(err)
	}

	items := []item{}

	for _, item := range bucket {
		b, err := t.request(item)
		if err != nil {
			panic(err)
		}

		i, err := t.parser(b)
		if err != nil {
			panic(err)
		}

		items = append(items, i)
	}

	fmt.Println(items)
}
