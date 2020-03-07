package main

//go get -u -v github.com/zmb3/gogetdoc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	//// gin.H 是 map[string]interface{} 的一种快捷方式
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	/// 使用 AsciiJSON 生成具有转义的非 ASCII 字符的 ASCII-only JSON
	r.GET("/asciiJson", func(c *gin.Context) {
		data := map[string]interface{}{
			"lang": "GO语言",
			"tag":  "<br>",
		}
		c.AsciiJSON(http.StatusOK, data)
	})

	/// HTML 渲染
	r.LoadHTMLGlob("templates/*")
	//r.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	/// 使用 JSONP 向不同域的服务器请求数据。如果查询参数存在回调，则将回调添加到响应体中。
	r.GET("/JSONP", func(c *gin.Context) {
		data := map[string]interface{}{
			"foo": "bar",
		}
		// /JSONP?callback=x
		// 将输出：x({\"foo\":\"bar\"})
		c.JSONP(http.StatusOK, data)
	})

	/// 通常，JSON 使用 unicode 替换特殊 HTML 字符，例如 < 变为 \ u003c。如果要按字面对这些字符进行编码，则可以使用 PureJSON。Go 1.6 及更低版本无法使用此功能。
	// 提供unicode实体
	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"html": "<b>Hello, wrold!</b>",
		})
	})
	// 提供字面字符
	r.GET("/purejson", func(c *gin.Context) {
		c.PureJSON(http.StatusOK, gin.H{
			"html": "<b>Hello, world!</b>",
		})
	})

	/// Query 和 post form
	// POST /post?id=1234&page=1 HTTP/1.1
	// Content-Type: application/x-www-form-urlencoded
	//
	// name=manu&message=this_is_great
	r.POST("/post", func(c *gin.Context) {
		id := c.Query("id")
		page := c.DefaultQuery("page", "0")
		name := c.PostForm("name")
		message := c.PostForm("message")
		fmt.Printf("id: %s; page: %s; name: %s; message: %s", id, page, name, message)
	})

	/// 使用 SecureJSON 防止 json 劫持。如果给定的结构是数组值，则默认预置 "while(1)," 到响应体。
	r.GET("/someJSON", func(c *gin.Context) {
		names := []string{"lena", "asutin", "foo"}
		c.SecureJSON(http.StatusOK, names)
	})

	/// 上传单文件
	r.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		fmt.Println(file.Filename)
		c.SaveUploadedFile(file, "./savefile")
		c.String(http.StatusOK, "upload succeed.")
	})

	/// 上传多文件
	r.POST("/uploads", func(c *gin.Context) {
		form, _ := c.MultipartForm()
		files := form.File["file[]"]
		for i, file := range files {
			c.SaveUploadedFile(file, "./savefile"+strconv.Itoa(i))
		}
		c.String(http.StatusOK, "uploads succeed.")
	})

	/// 从reader读取数据
	r.GET("/someDataFromReader", func(c *gin.Context) {
		response, err := http.Get("https://github.com/gongluck/gindemo")
		if err != nil || response.StatusCode != http.StatusOK {
			c.Status(http.StatusServiceUnavailable)
			return
		}
		reader := response.Body
		contentLength := response.ContentLength
		contentType := response.Header.Get("Content-Type")
		c.DataFromReader(http.StatusOK, contentLength, contentType, reader, nil)
	})

	/// 使用BasicAuth中间件
	authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
		"test1": "test11",
		"test2": "test22",
	}))
	authorized.GET("/authorized", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		fmt.Println("user : ", user)
		c.String(http.StatusOK, "BasicAuth.")
	})

	/// 使用HTTP方法
	r.GET("/someGet", defaulthttp)
	r.POST("/somePost", defaulthttp)
	r.PUT("/somePut", defaulthttp)
	r.DELETE("/someDelete", defaulthttp)
	r.PATCH("/somePatch", defaulthttp)
	r.HEAD("/someHead", defaulthttp)
	r.OPTIONS("/someOptions", defaulthttp)

	/// 优雅地重启或停止
	srv := http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		srv.ListenAndServe()
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	//r.Run(":8080")
}

func defaulthttp(c *gin.Context) {
	fmt.Println("http", c.Request.Method)
}
