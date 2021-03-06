module github.com/lazyfury/go-web-start/server

go 1.16

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.1
	github.com/gorilla/websocket v1.4.2
	github.com/lazyfury/go-web-template v1.0.23
	gorm.io/gorm v1.21.8
)

replace github.com/lazyfury/go-web-template => ../../go-web-template
