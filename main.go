package main

import (
	"net/http"

	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/gin-gonic/gin"
)

type Query struct {
	Url string `form:"url"`
}

func main() {
	r := gin.Default()

	r.GET("/open-graph", func(c *gin.Context) {
		var query Query
		if c.BindQuery(&query) != nil {
			c.JSON(400, gin.H{
				"code":    "QUERY_PARAM_MISSING",
				"message": "'url' is a required query param",
			})
			return
		}
		resp, err := http.Get(query.Url)

		if err != nil {
			c.JSON(500, gin.H{
				"code":                "INTERNA",
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

		c.Writer.Header().Add("Content-Type", "application/json")
		c.Writer.Write(json)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
