package handler

import (
	"chat_go/api/rpc"
	"chat_go/config"
	"chat_go/proto"
	"chat_go/tools"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strconv"
)

func Push(c *gin.Context) {
	var formPush FormPush
	if err := c.ShouldBindBodyWith(&formPush, binding.JSON); err != nil {
		tools.FailWithMsg(c, err.Error())
		return
	}
	authToken := formPush.AuthToken
	msg := formPush.Msg
	toUserId := formPush.ToUserId
	toUserIdInt, _ := strconv.Atoi(toUserId)
	getUserNameReq := &proto.GetUserInfoRequest{UserId: toUserIdInt}
	code, toUserName := rpc.RpcLogicObj.GetUserNameByUserId(getUserNameReq)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, "rpc fail get friend userName")
		return
	}
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: authToken}
	code, fromUserId, fromUserName := rpc.RpcLogicObj.CheckAuth(checkAuthReq)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, "rpc fail get self info")
		return
	}
	roomId := formPush.RoomId
	req := &proto.Send{
		Msg:          msg,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		ToUserId:     toUserIdInt,
		ToUserName:   toUserName,
		RoomId:       roomId,
		Op:           config.OpSingleSend,
	}
	code, rpcMsg := rpc.RpcLogicObj.Push(req)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, rpcMsg)
		return
	}
	tools.SuccessWithMsg(c, "ok", nil)
	return
}

func PushRoom(c *gin.Context) {
	var formRoom FormRoom
	if err := c.ShouldBindBodyWith(&formRoom, binding.JSON); err != nil {
		tools.FailWithMsg(c, err.Error())
		return
	}
	authToken := formRoom.AuthToken
	msg := formRoom.Msg
	roomId := formRoom.RoomId
	checkAuthReq := &proto.CheckAuthRequest{AuthToken: authToken}
	authCode, fromUserId, fromUserName := rpc.RpcLogicObj.CheckAuth(checkAuthReq)
	if authCode == tools.CodeFail {
		tools.FailWithMsg(c, "rpc fail get self info")
		return
	}
	req := &proto.Send{
		Msg:          msg,
		FromUserId:   fromUserId,
		FromUserName: fromUserName,
		RoomId:       roomId,
		Op:           config.OpRoomSend,
	}
	code, msg := rpc.RpcLogicObj.PushRoom(req)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, "rpc push room msg fail!")
		return
	}
	tools.SuccessWithMsg(c, "ok", msg)
	return
}

func Count(c *gin.Context) {
	var formCount FormCount
	if err := c.ShouldBindBodyWith(&formCount, binding.JSON); err != nil {
		tools.FailWithMsg(c, err.Error())
		return
	}
	roomId := formCount.RoomId
	req := &proto.Send{
		RoomId: roomId,
		Op:     config.OpRoomCountSend,
	}
	code, msg := rpc.RpcLogicObj.Count(req)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, "rpc get room count fail!")
		return
	}
	tools.SuccessWithMsg(c, "ok", msg)
	return
}

func GetRoomInfo(c *gin.Context) {
	var formRoomInfo FormRoomInfo
	if err := c.ShouldBindBodyWith(&formRoomInfo, binding.JSON); err != nil {
		tools.FailWithMsg(c, err.Error())
		return
	}
	roomId := formRoomInfo.RoomId
	req := &proto.Send{
		RoomId: roomId,
		Op:     config.OpRoomInfoSend,
	}
	code, msg := rpc.RpcLogicObj.GetRoomInfo(req)
	if code == tools.CodeFail {
		tools.FailWithMsg(c, "rpc get room info fail!")
		return
	}
	tools.SuccessWithMsg(c, "ok", msg)
	return
}
