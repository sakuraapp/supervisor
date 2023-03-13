package service

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sakuraapp/shared/pkg/model"
	"github.com/sakuraapp/shared/pkg/resource"
	"github.com/sakuraapp/shared/pkg/resource/opcode"
	"github.com/sakuraapp/shared/pkg/resource/permission"
	"strconv"
	"time"
)

const queueKey = "queue"

type Queue struct {
	app *Server
}

type QueueUpdateMessage struct {
	Active bool `json:"active"`
	Position int64 `json:"position"`
}

func (q *Queue) Add(ctx context.Context, roomId model.RoomId) error {
	strRoomId := strconv.FormatInt(int64(roomId), 10)
	pipe := q.app.GetRedis().Pipeline()

	pipe.ZAddNX(ctx, queueKey, &redis.Z{
		Score: float64(time.Now().Unix()),
		Member: strRoomId,
	})

	rankCmd := pipe.ZRank(ctx, queueKey, strRoomId)

	_, err := pipe.Exec(ctx)

	if err != nil {
		return err
	}

	notifMsg := resource.ServerMessage{
		Target: &resource.MessageTarget{
			Permissions: permission.MANAGE_ROOM,
		},
		Data: resource.BuildPacket(opcode.QueueUpdate, &QueueUpdateMessage{
			Active: true,
			Position: rankCmd.Val(),
		}),
	}

	err = q.app.DispatchRoom(roomId, &notifMsg)

	return nil
}

func (q *Queue) Pop(ctx context.Context) (model.RoomId, error) {
	el, err := q.app.GetRedis().ZPopMin(ctx, queueKey).Result()

	if err != nil {
		return 0, err
	}

	strRoomId := el[0].Member.(string)
	intRoomId, err := strconv.ParseInt(strRoomId, 10, 64)

	if err != nil {
		return 0, err
	}

	return model.RoomId(intRoomId), nil
}

func (q *Queue) Has(ctx context.Context, roomId model.RoomId) (bool, error) {
	strRoomId := strconv.FormatInt(int64(roomId), 10)
	_, err := q.app.GetRedis().ZRank(ctx, queueKey, strRoomId).Result()

	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (q *Queue) Size(ctx context.Context) (int64, error) {
	return q.app.GetRedis().ZCard(ctx, queueKey).Result()
}

func NewQueue(app *Server) *Queue {
	return &Queue{app: app}
}