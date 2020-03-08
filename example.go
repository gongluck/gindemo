package main

//go get -u -v github.com/zmb3/gogetdoc
//go get -u -v github.com/nsf/gocode

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	/// 强制日志颜色化
	gin.ForceConsoleColor()
	/// 禁止日志的颜色
	//gin.DisableConsoleColor()

	/// 记录日志
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	/// 定义路由日志的格式
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	//r := gin.Default()

	/// 使用中间件
	//新建一个没有任何默认中间件的路由
	r := gin.New()
	//全局中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(Logger())

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

	/// 只绑定 url 查询字符串
	type Person struct {
		Name    string `form:"name" uri:"name" binding:"required"`
		Address string `form:"address" uri:"address" binding:"required"`
	}
	var p Person
	r.Any("/ShouldBindQuery", func(c *gin.Context) {
		if c.ShouldBindQuery(&p) == nil {
			fmt.Println("====== Only Bind By Query String ======")
			fmt.Println(p.Name)
			fmt.Println(p.Address)
		}
		c.String(http.StatusOK, "Success")
	})

	/// 当在中间件或 handler 中启动新的 Goroutine 时，不能使用原始的上下文，必须使用只读副本。
	r.GET("/long_async", func(c *gin.Context) {
		cc := c.Copy() //应该类似于增加引用计数
		go func() {
			time.Sleep(5 * time.Second)
			fmt.Println("Done in path" + cc.Request.URL.Path)
		}()
		c.String(http.StatusOK, "Success")
	})

	/// 绑定URL
	r.GET("/bindingurl/:name/:address", func(c *gin.Context) {
		type Person struct {
			Name    string `form:"name" uri:"name" binding:"required"`
			Address string `form:"address" uri:"address" binding:"required"`
		}
		var p Person
		if err := c.ShouldBindUri(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": p.Name, "address": p.Address})
	})

	/// 重定向
	r.GET("/redirect", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://github.com/gongluck/gindemo.git")
	})
	r.GET("/redirect2", func(c *gin.Context) {
		c.Request.URL.Path = "/redirect"
		r.HandleContext(c)
	})

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

/// 自定义中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// 设置变量
		c.Set("mymiddleware", "gongluck")
		// 请求前
		c.Next()
		// 请求后
		latency := time.Since(t)
		log.Print(latency)
		// 获取发送的 status
		status := c.Writer.Status()
		log.Println(status)
	}
}
