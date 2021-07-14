package router

import (
	"chat_go/api/rpc"
	"chat_go/proto"
	"chat_go/tools"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		var openCorsFlag = true
		if openCorsFlag {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			c.Header("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, nil)
		}
		c.Next()
	}
}


func CheckSessionId() gin.HandlerFunc {
	return func(c *gin.Context) {
		var formCheckSessionId FormCheckSessionId
		if err := c.ShouldBindBodyWith(&formCheckSessionId, binding.JSON); err != nil {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, nil, nil)
			return
		}
		authToken := formCheckSessionId.AuthToken
		req := &proto.CheckAuthRequest{
			AuthToken: authToken,
		}
		code, userId, userName := rpc.RpcLogicObj.CheckAuth(req)
		if code == tools.CodeFail || userId <= 0 || userName == "" {
			c.Abort()
			tools.ResponseWithCode(c, tools.CodeSessionError, nil, nil)
			return
		}
		c.Next()
		return
	}
}
