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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-zookeeper/zk"
	"github.com/kitex-contrib/registry-zookeeper/entity"
	"github.com/kitex-contrib/registry-zookeeper/utils"
)

type zookeeperRegistry struct {
	conn           *zk.Conn
	authOpen       bool
	user, password string
}

func NewZookeeperRegistry(servers []string, sessionTimeout time.Duration) (registry.Registry, error) {
	con, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperRegistry{conn: con}, nil
}

func NewZookeeperRegistryWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (registry.Registry, error) {
	if user == "" || password == "" {
		return nil, fmt.Errorf("user or password can't be empty")
	}
	con, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	auth := []byte(fmt.Sprintf("%s:%s", user, password))
	err = con.AddAuth(utils.Scheme, auth)
	if err != nil {
		return nil, err
	}
	return &zookeeperRegistry{conn: con, authOpen: true, user: user, password: password}, nil
}

func (z *zookeeperRegistry) Register(info *registry.Info) error {
	path, err := buildPath(info)
	if err != nil {
		return err
	}
	content, err := json.Marshal(&entity.RegistryEntity{Weight: info.Weight, Tags: info.Tags})
	if err != nil {
		return err
	}
	return z.createNode(path, content, true)
}

/**
  path format as follow:
  /{ serviceName}/{ip}:{port}
*/
func buildPath(info *registry.Info) (string, error) {
	var path string
	if info == nil {
		return "", fmt.Errorf("registry info can't be nil")
	}
	if info.ServiceName == "" {
		return "", fmt.Errorf("registry info service name can't be empty")
	}
	if info.Addr == nil {
		return "", fmt.Errorf("registry info addr can't be nil")
	}
	if !strings.HasPrefix(info.ServiceName, utils.Separator) {
		path = utils.Separator + info.ServiceName
	}

	if host, port, err := net.SplitHostPort(info.Addr.String()); err == nil {
		if port == "" {
			return "", fmt.Errorf("registry info addr missing port")
		}
		if host == "" {
			ipv4, err := utils.GetLocalIPv4Address()
			if err != nil {
				return "", fmt.Errorf("get local ipv4 error, cause %w", err)
			}
			path = path + utils.Separator + ipv4 + ":" + port
		} else {
			path = path + utils.Separator + host + ":" + port
		}
	} else {
		return "", fmt.Errorf("parse registry info addr error")
	}
	return path, nil
}

func (z *zookeeperRegistry) Deregister(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("registry info can't be nil")
	}
	path, err := buildPath(info)
	if err != nil {
		return err
	}
	return z.deleteNode(path)
}

func (z *zookeeperRegistry) createNode(path string, content []byte, ephemeral bool) error {
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
	if z.authOpen {
		_, err := z.conn.Create(path, content, flag, zk.DigestACL(zk.PermAll, z.user, z.password))
		if err != nil {
			return fmt.Errorf("create node [%s] with auth error, cause %w", path, err)
		}
		return nil
	} else {
		_, err := z.conn.Create(path, content, flag, zk.WorldACL(zk.PermAll))
		if err != nil {
			return fmt.Errorf("create node [%s] error, cause %w", path, err)
		}
		return nil
	}
}

func (z *zookeeperRegistry) deleteNode(path string) error {
	err := z.conn.Delete(path, -1)
	if err != nil {
		return fmt.Errorf("delete node [%s] error, cause %w", path, err)
	}
	return nil
}
