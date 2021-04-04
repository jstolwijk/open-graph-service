package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/allegro/bigcache"
	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
	"github.com/gin-gonic/gin"
)

type Query struct {
	Url string `form:"url"`
}

var cacheManager cache.Cache
var httpClient http.Client

func main() {
	r := gin.Default()

	bigcacheClient, _ := bigcache.NewBigCache(bigcache.DefaultConfig(4 * time.Hour))
	bigcacheStore := store.NewBigcache(bigcacheClient, nil) // No otions provided (as second argument)

	cacheManager := cache.New(bigcacheStore)

	// TODO fix to prevent man in the middle
	// This hack is needed to get the docker image working, cosider making a docker image with openssl to fix
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = http.Client{Transport: customTransport}

	r.GET("/open-graph", func(c *gin.Context) {
		var query Query
		if c.BindQuery(&query) != nil {
			c.JSON(400, gin.H{
				"code":    "QUERY_PARAM_MISSING",
				"message": "'url' is a required query param",
			})
			return
		}

		value, err := cacheManager.Get(query.Url)

		if err == nil {
			b, ok := value.([]byte)

			if !ok {
				c.JSON(500, gin.H{
					"code":                "2",
					"origionalStatusCode": "asd",
					"message":             err,
				})
				return
			}

			fmt.Printf("Cache hit\n")
			c.Writer.Header().Add("Content-Type", "application/json")
			c.Writer.Write(b)
			return
		}

		// TODO validate if response status code is 200
		resp, err := httpClient.Get(query.Url)

		if err != nil {
			c.JSON(500, gin.H{
				"code":                "3",
				"origionalStatusCode": "asd",
				"message":             err,
			})
			return
		}
		defer resp.Body.Close()
		// body, err := ioutil.ReadAll(resp.Body)

		// if err != nil {
		// 	c.JSON(500, gin.H{
		// 		"code":                "INTERNA",
		// 		"origionalStatusCode": "asd",
		// 		"message":             err,
		// 	})
		// 	return
		// }

		og := opengraph.NewOpenGraph()

		err = og.ProcessHTML(resp.Body)

		json, err := og.ToJSON()

		err = cacheManager.Set(query.Url, json, nil) // TODO set options

		c.Writer.Header().Add("Content-Type", "application/json")
		c.Writer.Write(json)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
