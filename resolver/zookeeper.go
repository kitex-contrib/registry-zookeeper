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

package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/go-zookeeper/zk"
	"github.com/kitex-contrib/registry-zookeeper/entity"
	"github.com/kitex-contrib/registry-zookeeper/utils"
)

type zookeeperResolver struct {
	con *zk.Conn
}

// NewZookeeperResolver create a zookeeper based resolver
func NewZookeeperResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	con, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{con: con}, nil
}

// NewZookeeperResolver create a zookeeper based resolver with auth
func NewZookeeperResolverWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (discovery.Resolver, error) {
	con, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	auth := []byte(fmt.Sprintf("%s:%s", user, password))
	err = con.AddAuth(utils.Scheme, auth)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{con: con}, nil
}

// Target implements the Resolver interface.
func (r *zookeeperResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) string {
	return target.ServiceName()
}

// Resolve implements the Resolver interface.
func (r *zookeeperResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	if !strings.HasPrefix(desc, utils.Separator) {
		desc = utils.Separator + desc
	}
	eps, err := r.getEndPoints(desc)
	if err != nil {
		return discovery.Result{}, err
	}
	instances, err := r.getInstances(eps, desc)
	if err != nil {
		return discovery.Result{}, err
	}
	res := discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}
	return res, nil
}

func (r *zookeeperResolver) getEndPoints(name string) ([]string, error) {
	child, _, err := r.con.Children(name)
	return child, err
}

func (r *zookeeperResolver) detailEndPoints(service, ep string) (discovery.Instance, error) {
	data, _, err := r.con.Get(service + utils.Separator + ep)
	if err != nil {
		return nil, err
	}
	en := new(entity.RegistryEntity)
	err = json.Unmarshal(data, en)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data [%s] error, cause %w", data, err)
	}
	return discovery.NewInstance("tcp", ep, en.Weight, en.Tags), nil
}

func (r *zookeeperResolver) getInstances(eps []string, service string) ([]discovery.Instance, error) {
	instances := make([]discovery.Instance, 0, len(eps))
	for _, ep := range eps {
		if host, port, err := net.SplitHostPort(ep); err == nil {
			if port == "" {
				return []discovery.Instance{}, fmt.Errorf("missing port when parse node [%s]", ep)
			}
			if host == "" {
				return []discovery.Instance{}, fmt.Errorf("missing host when parse node [%s]", ep)
			}
			ins, err := r.detailEndPoints(service, ep)
			if err != nil {
				return []discovery.Instance{}, fmt.Errorf("detail endpoint [%s] info error, cause %w", ep, err)
			}
			instances = append(instances, ins)
		} else {
			return []discovery.Instance{}, fmt.Errorf("parse node [%s] error, details info [%w]", ep, err)
		}
	}
	return instances, nil
}

// Diff implements the Resolver interface.
func (r *zookeeperResolver) Diff(key string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(key, prev, next)
}

// Name implements the Resolver interface.
func (r *zookeeperResolver) Name() string {
	return "zookeeper"
}
