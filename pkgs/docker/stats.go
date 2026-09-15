package docker

import (
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/rehiy/libgo/logman"
)

// ContainerStats 获取 Docker SDK 原始容器统计。
func (s *DockerService) ContainerStats(ctx context.Context, id string) (container.StatsResponse, error) {
	stats, err := s.client.ContainerStats(ctx, id, false)
	if err != nil {
		logman.Error("Get container stats failed", "id", id, "error", err)
		return container.StatsResponse{}, err
	}
	defer stats.Body.Close()

	data, err := io.ReadAll(stats.Body)
	if err != nil {
		logman.Error("Read container stats failed", "id", id, "error", err)
		return container.StatsResponse{}, err
	}

	var v container.StatsResponse
	if err := json.Unmarshal(data, &v); err != nil {
		logman.Error("Parse container stats failed", "id", id, "error", err)
		return container.StatsResponse{}, err
	}
	return v, nil
}

// ContainerTop 获取容器内进程列表；精简镜像内无 ps 时会失败，由调用方决定是否降级。
func (s *DockerService) ContainerTop(ctx context.Context, id string) (*container.TopResponse, error) {
	top, err := s.client.ContainerTop(ctx, id, nil)
	if err != nil {
		return nil, err
	}
	return &top, nil
}
