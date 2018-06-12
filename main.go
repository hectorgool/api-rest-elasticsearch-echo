package main

import (
	"github.com/hectorgool/api-rest-elasticsearch-echo/common"
	"github.com/hectorgool/api-rest-elasticsearch-echo/elasticsearch"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

/*
type Data struct {
  data interface{} `json:"data"`
}
*/

func main() {

	defer common.Logfile.Close()

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:9000"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.GET("/", func(c echo.Context) error {
		result, err := elasticsearch.Ping()
		common.CheckError(err)
		return c.String(http.StatusOK, result)
	})

	e.GET("/search", func(c echo.Context) error {
		q := c.QueryParam("q")
		result, err := elasticsearch.Search(q)
		common.CheckError(err)
		return c.JSONPretty(http.StatusOK, result, "  ")
	})

	//6a9f72f5-eb28-4be4-a9d2-8f328d938123
	e.GET("/search/:id", func(c echo.Context) error {
		id := c.Param("id")
		doc := elasticsearch.ReadDocument(id)
		return c.JSONPretty(http.StatusOK, doc, "  ")
	})

	e.Logger.Fatal(e.Start(":8080"))

}
