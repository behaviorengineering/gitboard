// Package benchnet provides network call-count and latency benchmarks.
package benchnet

import (
	"context"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/pkg/boardconfig"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
	"github.com/behaviorengineering/gitboard/pkg/remotegit/remotegittest"
)

func BenchmarkCollectWarmGitHub(b *testing.B) {
	run := remotegittest.ConformityGitHubExec()
	svc := dashboard.DefaultServiceConfig(run).CreateService()
	heads := 600
	merged := 600
	doc := boardconfig.File{
		Upstream: boardconfig.Upstream{HeadsSeconds: &heads, MergedSeconds: &merged},
		Projects: make([]boardconfig.Project, 8),
	}
	for i := range doc.Projects {
		doc.Projects[i] = boardconfig.Project{
			ID: "p", Label: "P", Host: boardconfig.HostGitHub, Path: "acme/app",
		}
		doc.Projects[i].ID = "p" + string(rune('a'+i))
	}
	ctx := context.Background()
	if _, err := svc.Collect(ctx, doc, false, ""); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		if _, err := svc.Collect(ctx, doc, false, ""); err != nil {
			b.Fatal(err)
		}
		_ = time.Since(start)
	}
}

func BenchmarkCollectColdGitHub(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		run := remotegittest.ConformityGitHubExec()
		svc := dashboard.DefaultServiceConfig(run).CreateService()
		doc := boardconfig.File{
			Projects: []boardconfig.Project{{
				ID: "p1", Label: "P1", Host: boardconfig.HostGitHub, Path: "acme/app",
			}},
		}
		ctx := context.Background()
		b.StartTimer()
		if _, err := svc.Collect(ctx, doc, true, ""); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHeadsSingleflight(b *testing.B) {
	cache := remotegit.NewTTLCache(time.Now)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := cache.GetOrLoadHeads("bench/key", time.Minute, false, func() (remotegit.HeadsSnapshot, error) {
				time.Sleep(time.Millisecond)
				return remotegit.HeadsSnapshot{DefaultBranch: "main"}, nil
			})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
