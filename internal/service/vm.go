package service

import (
	"context"
	"github.com/sakuraapp/shared/pkg/model"
	"github.com/sakuraapp/supervisor/internal/supervisor"
)

type VMProvider interface {
	Deploy(ctx context.Context, roomId model.RoomId, region supervisor.Region) error
	Destroy(ctx context.Context, roomId model.RoomId) error
	GetAvailableRooms(ctx context.Context, region supervisor.Region) (int64, error)
}

type DeployOptions struct {
	RoomId model.RoomId
	Region supervisor.Region

}