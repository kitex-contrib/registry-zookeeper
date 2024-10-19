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
	"time"

	zookresolver "github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/zookeeperkitex/resolver"
	"github.com/cloudwego/kitex/pkg/discovery"
)

// NewZookeeperResolver create a zookeeper based resolver
func NewZookeeperResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	return zookresolver.NewZookeeperResolver(servers, sessionTimeout)
}

// NewZookeeperResolver create a zookeeper based resolver with auth
func NewZookeeperResolverWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (discovery.Resolver, error) {
	return zookresolver.NewZookeeperResolverWithAuth(servers, sessionTimeout, user, password)
}
