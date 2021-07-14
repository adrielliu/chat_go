package connect

import (
	"chat_go/proto"
	"sync"
	"sync/atomic"
)

type BucketOptions struct {
	ChannelSize   int
	RoomSize      int
	RoutineAmount uint64
	RoutineSize   int
}

type Bucket struct {
	bucketOptions BucketOptions
	cLock         sync.RWMutex
	chs           map[int]*UserChannel             // user conns
	rooms         map[int]*Room                    // bucket rooms
	routines      []chan *proto.PushRoomMsgRequest // msg chans
	routinesNum   uint64
	broadcast     chan []byte
}

func NewBucket(bucketOpt BucketOptions) (b *Bucket) {
	b = new(Bucket)
	b.bucketOptions = bucketOpt
	b.chs = make(map[int]*UserChannel, bucketOpt.ChannelSize)
	b.routines = make([]chan *proto.PushRoomMsgRequest, bucketOpt.RoutineAmount)
	b.rooms = make(map[int]*Room, bucketOpt.RoomSize)
	for i := uint64(0); i < b.bucketOptions.RoutineAmount; i++ {
		c := make(chan *proto.PushRoomMsgRequest, bucketOpt.RoutineSize)
		b.routines[i] = c
		go b.PushRoomMSG(c)
	}
	return
}

func (b *Bucket) PushRoomMSG(ch chan *proto.PushRoomMsgRequest) {
	for {
		var (
			msg  *proto.PushRoomMsgRequest
			room *Room
		)
		msg = <-ch
		if room = b.GetRoomInfoByID(msg.RoomId); room != nil{
			room.SendRoomMessage(&msg.Msg)
		}
	}
}

func (b *Bucket) GetRoomInfoByID(rid int) (room *Room) {
	b.cLock.RLock()
	defer b.cLock.RUnlock()
	room, _ = b.rooms[rid]
	return
}

func (b *Bucket) AddChannel (userId int, roomId int, ch *UserChannel) (err error) {
	// new room  or  new channel
	var(
		room *Room
		ok bool
	)
	b.cLock.Lock()
	if roomId != NoRoom{
		if room, ok = b.rooms[roomId]; !ok{
			room = NewRoom(roomId)
			b.rooms[roomId] = room
		}
		ch.Room = room
	}
	ch.userId = userId
	b.chs[userId] = ch
	b.cLock.Unlock()
	if room != nil{
		err = room.AddUser(ch)
	}
	return
}

func (b *Bucket) DeleteChannel(ch *UserChannel)  {
	var (
		ok bool
		room *Room
	)
	b.cLock.Lock()
	defer b.cLock.Unlock()
	if ch, ok = b.chs[ch.userId]; ok{
		room = b.chs[ch.userId].Room
		// delete from bucket
		delete(b.chs, ch.userId)
	}
	if room != nil && room.DeleteUser(ch){
		// if room empty delete,will mark room.drop is true
		if room.drop == true{
			delete(b.rooms, room.Id)
		}
	}

}

func (b *Bucket) GetChannelByID(userId int) (ch *UserChannel) {
	b.cLock.RLock()
	ch = b.chs[userId]
	b.cLock.RUnlock()
	return
}

func (b *Bucket) BroadcastRoom(req *proto.PushRoomMsgRequest)  {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.bucketOptions.RoutineAmount
	b.routines[num] <- req
}