// Copyright 2021 CloudWeGo authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific

package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-zookeeper/zk"
	"github.com/kitex-contrib/registry-zookeeper/entity"
	"github.com/kitex-contrib/registry-zookeeper/utils"
)

var (
	ErrorZkConnectedTimedOut = errors.New("timed out waiting for zk connected")
	ErrorNilRegistryInfo     = errors.New("registry info can't be nil")
)

type zookeeperRegistry struct {
	sync.RWMutex
	conn           *zk.Conn
	user           string
	password       string
	sessionTimeout time.Duration
	canceler       map[string]context.CancelFunc
}

func NewZookeeperRegistry(servers []string, sessionTimeout time.Duration) (registry.Registry, error) {
	return NewZookeeperRegistryWithAuth(servers, sessionTimeout, "", "")
}

func NewZookeeperRegistryWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (registry.Registry, error) {
	conn, event, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	if user != "" && password != "" {
		if err := conn.AddAuth(utils.Scheme, []byte(fmt.Sprintf("%s:%s", user, password))); err != nil {
			return nil, err
		}
	}
	ticker := time.NewTimer(sessionTimeout / 2)
	for {
		select {
		case e := <-event:
			if e.State == zk.StateConnected {
				return &zookeeperRegistry{
					user:           user,
					password:       password,
					sessionTimeout: sessionTimeout,
					conn:           conn,
					canceler:       make(map[string]context.CancelFunc),
				}, nil
			}
		case <-ticker.C:
			return nil, ErrorZkConnectedTimedOut
		}
	}
}

func (z *zookeeperRegistry) Register(info *registry.Info) error {
	if info == nil {
		return ErrorNilRegistryInfo
	}
	ne := entity.MustNewNodeEntity(info)
	path, err := ne.Path()
	if err != nil {
		return err
	}
	content, err := ne.Content()
	if err != nil {
		return err
	}
	err = z.createNode(path, content, true)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	z.Lock()
	defer z.Unlock()
	z.canceler[path] = cancel
	go z.keepalive(ctx, path, content)
	return nil
}

func (z *zookeeperRegistry) Deregister(info *registry.Info) error {
	if info == nil {
		return ErrorNilRegistryInfo
	}
	ne := entity.MustNewNodeEntity(info)
	path, err := ne.Path()
	if err != nil {
		return err
	}
	z.Lock()
	defer z.Unlock()
	cancel, ok := z.canceler[path]
	if ok {
		cancel()
		delete(z.canceler, path)
	}
	return z.deleteNode(path)
}

func (z *zookeeperRegistry) createNode(path string, content []byte, ephemeral bool) error {
	exists, stat, err := z.conn.Exists(path)
	if err != nil {
		return err
	}
	// ephemeral nodes handling after restart
	// fixes a race condition if the server crashes without using CreateProtectedEphemeralSequential()
	// https://github.com/go-kratos/kratos/blob/main/contrib/registry/zookeeper/register.go
	if exists && ephemeral {
		err = z.conn.Delete(path, stat.Version)
		if err != nil && err != zk.ErrNoNode {
			return err
		}
		exists = false
	}
	if !exists {
		i := strings.LastIndex(path, utils.Separator)
		if i > 0 {
			err := z.createNode(path[0:i], nil, false)
			if err != nil && !errors.Is(err, zk.ErrNodeExists) {
				return err
			}
		}
		var flag int32
		if ephemeral {
			flag = zk.FlagEphemeral
		}
		if z.user != "" && z.password != "" {
			_, err = z.conn.Create(path, content, flag, zk.DigestACL(zk.PermAll, z.user, z.password))
		} else {
			_, err = z.conn.Create(path, content, flag, zk.WorldACL(zk.PermAll))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (z *zookeeperRegistry) deleteNode(path string) error {
	return z.conn.Delete(path, -1)
}

func (z *zookeeperRegistry) keepalive(ctx context.Context, path string, content []byte) {
	sessionID := z.conn.SessionID()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cur := z.conn.SessionID()
			if cur != 0 && sessionID != cur {
				if err := z.createNode(path, content, true); err == nil {
					sessionID = cur
				}
			}
		}
	}
}
