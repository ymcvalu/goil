package main

import (
	"fmt"
	"goil"
	"goil/session"
	"net/http"
	"os"
)

type Params struct {
	Name      string   `form:"name" validator:"reg(^[a-z]*$)"`
	PName     *string  `form:"name"`
	PAge      *int     `form:"age" validator:"range(0 150) min(0) max(150)"`
	Age       int      `form:"age"`
	FileName  string   `file:"music"`
	PFileName string   `file:"music"`
	Size      int64    `file:"music"`
	PSize     *int64   `file:"music"`
	FileF     os.File  `file:"music"`
	PFileF    *os.File `file:"music"`
	//File     os.File `file:"music"`
	Path *struct {
		Path      string       `path:"path"`
		PPath     *string      `path:"path"`
		PPPtrPath ******string `path:"path"`
		Slice     []string     `form:"slice"`
		PSlice    *[]string    `form:"slice"`
	}
	File  goil.File  `file:"music"`
	PFile *goil.File `file:"music"`
}

func main() {
	session.EnableRdsSession()
	app := goil.Default()

	app.GET("/", func(c *goil.Context) {
		c.Text("hello,goil!")
	})
	goil.HtmlTemp("index", "./index.html")
	app.GET("/json", func(c *goil.Context) {
		c.JSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
	})
	app.GET("/indentJson", func(c *goil.Context) {
		c.IndentJSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
	})
	app.GET("/securyJson", func(c *goil.Context) {
		c.SecureJSON(map[string]string{
			"name": "Jim",
			"age":  "19",
		})
		query := c.Request.URL.Query()
		fmt.Println(query["get"])
	})
	app.POST("/json/echo/:path", func(c *goil.Context) {

		var params = Params{}
		err := c.Bind(&params)

		if err != nil {
			c.JSON(map[string]string{
				"Msg": err.Error(),
			})
			return
		}
		c.JSON(params)
		//c.Info(isatty.IsTerminal(os.Stdin.Fd()))
		c.Info(params.FileF, params.PFileF, params.File.File, params.PFile.File)
	})
	app.GET("/param/:name/:age", func(c *goil.Context) {
		if val := c.Param("age"); val != "" {
			c.Text(val)
		}
	})
	app.GET("/session", session.UseSession(), func(c *goil.Context) {
		sess, ok := c.Session.Get("sess").(string)
		if ok {
			c.Text(sess)
		} else {
			c.Text("sess not set")
		}
		c.Info("log")
	})
	app.POST("/session/:sess", session.UseSession(), func(c *goil.Context) {
		val := c.Param("sess")
		c.Session.Set("sess", val)
		c.Text("set")
		c.Text(c.Query("q"))
	})
	app.PUT("/rewrite_code", func(c *goil.Context) {
		c.Status(404)
	})
	app.GET("/index", goil.EnableGzip(goil.GZIP_DefaultCompression), func(c *goil.Context) {
		c.Html("index", struct{ Title, Content string }{
			Title:   "this is a title",
			Content: "this is the conent",
		})
	})
	app.GET("/file/*file", func(c *goil.Context) {
		c.File(c.Param("file"))
	})
	app.GET("/redirect", func(c *goil.Context) {
		c.Redirect(302, "http://www.baidu.com")
	})
	app.Static("/static", "")
	app.GET("/reverser_proxy", goil.ReverseProxy(func(r *http.Request) {
		u := r.URL
		u.Scheme = "http"
		u.Host = "127.0.0.1:8081"
		u.Path = "/file/bz1.jpeg"
	}))
	app.GET("/panic", func(c *goil.Context) {
		c.Panic("there are some error")
	})

	if err := app.Run(":8081"); err != nil {
		panic(err)
	}

}

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type Params struct {
// 	Name  string   `form:"name"`
// 	Age   *int     `form:"age"`
// 	Slice []string `form:"slice"`
// }

// func main() {

// 	router := gin.Default()

// 	router.POST("/", func(c *gin.Context) {
// 		p := Params{}
// 		c.Bind(&p)
// 		c.JSON(http.StatusOK, p)
// 	})
// 	router.Run(":8000")
// }
