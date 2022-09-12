package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var serviceMap map[string]string

func sessionProxy(c *gin.Context) {

	session := sessions.Default(c)

	var id uuid.UUID
	v := session.Get("uuid")
	if v == nil {
		id = uuid.New()
		session.Set("uuid", id.String())
		v = id.String()
	}
	session.Save()

	sessionId := v.(string)

	if _, ok := serviceMap[sessionId]; !ok {

		// if len(serviceMap)%2 == 0 {
		// 	url := "http://192.168.0.152:8002"
		// 	serviceMap[sessionId] = url
		// } else {
		// 	url := "http://192.168.0.152:9002"
		// 	serviceMap[sessionId] = url
		// }
		url := "https://bitnodes.io/nodes/live-map/"
		serviceMap[sessionId] = url
	}

	proxy(c, serviceMap[sessionId])
}

// func rewriteBody(resp *http.Response) (err error) {
// 	b, err := ioutil.ReadAll(resp.Body) //Read html
// 	if err != nil {
// 		return err
// 	}
// 	err = resp.Body.Close()
// 	if err != nil {
// 		return err
// 	}
// 	b = bytes.Replace(b, []byte("Top user agents"), []byte("客户端排行"), -1) // replace html
// 	body := ioutil.NopCloser(bytes.NewReader(b))
// 	resp.Body = body
// 	resp.ContentLength = int64(len(b))
// 	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

// 	// 去掉无用的header
// 	resp.Header.Del("Content-Security-Policy")
// 	resp.Header.Del("X-Frame-Options")
// 	return nil
// }

func rewriteHeader(resp *http.Response) (err error) {

	resp.Header.Del("Content-Security-Policy")
	resp.Header.Del("X-Frame-Options")
	return nil
}

func proxy(c *gin.Context, service string) {
	remote, err := url.Parse(service)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	//Define the director func
	//This is a good place to log, for example
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("proxyPath")
	}

	proxy.ModifyResponse = rewriteHeader

	proxy.ServeHTTP(c.Writer, c.Request)

}

func main() {

	serviceMap = make(map[string]string)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// // 加入html静态页面
	r.Use(static.Serve("/static", static.LocalFile("static", false)))

	// 404 处理
	r.NoRoute(func(c *gin.Context) {
		c.File("static/index.html")
	})
	//r.GET("/nodes/live-map", )

	//Create a catchall route
	r.Any("/*proxyPath", sessionProxy)

	r.Run(":8180")
}
