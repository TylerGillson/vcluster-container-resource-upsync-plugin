package main

import (
	"github.com/TylerGillson/vcluster-container-resource-upsync/syncers"
	"github.com/TylerGillson/vcluster-sdk/plugin"
)

func main() {
	ctx := plugin.MustInit()
	plugin.MustRegister(syncers.NewContainerResourceSyncer(ctx))
	plugin.MustStart()
}
