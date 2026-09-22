// Package benchnet provides network call-count and latency benchmarks.
package benchnet

import (
	"context"
	"testing"
	"time"

	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/pkg/dashboard"
	"github.com/behaviorengineering/gitboard/pkg/remotegit"
)

func BenchmarkCollectWarmGitHub(b *testing.B) {
	run := remotegit.ConformityGitHubExec()
	svc := dashboard.New(remotegit.NewGitHub(run), nil, nil, nil, nil)
	heads := 600
	merged := 600
	doc := config.File{
		Upstream: config.Upstream{HeadsSeconds: &heads, MergedSeconds: &merged},
		Projects: make([]config.Project, 8),
	}
	for i := range doc.Projects {
		doc.Projects[i] = config.Project{
			ID: "p", Label: "P", Host: config.HostGitHub, Path: "acme/app",
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
		run := remotegit.ConformityGitHubExec()
		svc := dashboard.New(remotegit.NewGitHub(run), nil, nil, nil, nil)
		doc := config.File{
			Projects: []config.Project{{
				ID: "p1", Label: "P1", Host: config.HostGitHub, Path: "acme/app",
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
	cache := remotegit.NewTTLCache()
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
