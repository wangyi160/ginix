package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

		if len(serviceMap)%2 == 0 {
			url := "http://192.168.0.152:8002"
			serviceMap[sessionId] = url
		} else {
			url := "http://192.168.0.152:9002"
			serviceMap[sessionId] = url
		}
	}

	proxy(c, serviceMap[sessionId])
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

	proxy.ServeHTTP(c.Writer, c.Request)
}

func main() {

	serviceMap = make(map[string]string)

	r := gin.Default()

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	//Create a catchall route
	r.Any("/*proxyPath", sessionProxy)

	r.Run(":8080")
}
