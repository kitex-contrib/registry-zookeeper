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
	"errors"
	zookregistry "github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/zookeeperkitex/registry"
	"github.com/cloudwego/kitex/pkg/registry"
	"time"
)

var (
	ErrorZkConnectedTimedOut = errors.New("timed out waiting for zk connected")
	ErrorNilRegistryInfo     = errors.New("registry info can't be nil")
)

func NewZookeeperRegistry(servers []string, sessionTimeout time.Duration) (registry.Registry, error) {
	return zookregistry.NewZookeeperRegistry(servers, sessionTimeout)
}

func NewZookeeperRegistryWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (registry.Registry, error) {
	return zookregistry.NewZookeeperRegistryWithAuth(servers, sessionTimeout, user, password)
}
