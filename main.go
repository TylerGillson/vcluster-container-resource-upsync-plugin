package main

import (
	"github.com/loft-sh/vcluster-sdk/plugin"
	"github.com/spectrocloud/vcluster-container-resource-upsync/syncers"
)

func main() {
	ctx := plugin.MustInit()
	plugin.MustRegister(syncers.NewContainerResourceSyncer(ctx))
	plugin.MustStart()
}
