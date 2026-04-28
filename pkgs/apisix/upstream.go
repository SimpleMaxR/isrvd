package apisix

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Upstream Apisix Upstream 信息
type Upstream struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Type       string `json:"type"`
	Nodes      any    `json:"nodes,omitempty"`
	HashOn     string `json:"hash_on,omitempty"`
	Key        string `json:"key,omitempty"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// ListUpstreams 获取所有 Upstream 列表
func (c *Client) ListUpstreams() ([]Upstream, error) {
	data, err := c.doRequest(http.MethodGet, "/upstreams", nil)
	if err != nil {
		return nil, err
	}
	return parseUpstreamList(data)
}

// GetUpstream 获取单个 Upstream 详情，保留 APISIX 原始配置结构
func (c *Client) GetUpstream(upstreamID string) (map[string]any, error) {
	data, err := c.doRequest(http.MethodGet, "/upstreams/"+upstreamID, nil)
	if err != nil {
		return nil, err
	}
	return parseSingleUpstream(data)
}

// CreateUpstream 创建 Upstream，透传 APISIX 配置
func (c *Client) CreateUpstream(req map[string]any) (map[string]any, error) {
	data, err := c.doRequest(http.MethodPost, "/upstreams", cleanUpstreamBody(req))
	if err != nil {
		return nil, err
	}
	return parseSingleUpstream(data)
}

// UpdateUpstream 更新 Upstream，透传 APISIX 配置
func (c *Client) UpdateUpstream(upstreamID string, req map[string]any) (map[string]any, error) {
	data, err := c.doRequest(http.MethodPut, "/upstreams/"+upstreamID, cleanUpstreamBody(req))
	if err != nil {
		return nil, err
	}
	return parseSingleUpstream(data)
}

// DeleteUpstream 删除 Upstream
func (c *Client) DeleteUpstream(upstreamID string) error {
	_, err := c.doRequest(http.MethodDelete, "/upstreams/"+upstreamID, nil)
	return err
}

// parseUpstreamList 解析 Apisix Upstream 列表响应
func parseUpstreamList(data []byte) ([]Upstream, error) {
	var raw struct {
		List []struct {
			Value Upstream `json:"value"`
		} `json:"list"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 Upstream 列表失败: %w", err)
	}
	result := make([]Upstream, 0, len(raw.List))
	for _, item := range raw.List {
		result = append(result, item.Value)
	}
	return result, nil
}

// parseSingleUpstream 解析单个 Upstream 响应
func parseSingleUpstream(data []byte) (map[string]any, error) {
	var raw struct {
		Value map[string]any `json:"value"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("解析 Upstream 详情失败: %w", err)
	}
	return raw.Value, nil
}

// cleanUpstreamBody 移除 APISIX 只读字段，其余字段原样透传
func cleanUpstreamBody(req map[string]any) map[string]any {
	body := make(map[string]any, len(req))
	for key, value := range req {
		body[key] = value
	}
	delete(body, "id")
	delete(body, "create_time")
	delete(body, "update_time")
	return body
}
