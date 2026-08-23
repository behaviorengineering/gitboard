// Local dev API: GitLab + GitHub project dashboard (127.0.0.1 only).
package main

import (
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/behaviorengineering/gitboard/internal/cliexec"
	"github.com/behaviorengineering/gitboard/internal/config"
	"github.com/behaviorengineering/gitboard/internal/dashboard"
	"github.com/behaviorengineering/gitboard/internal/forge"
	"github.com/behaviorengineering/gitboard/internal/server"
	"github.com/behaviorengineering/gitboard/internal/triage"
)

func main() {
	projectsPath := flag.String("projects", config.DefaultProjectsPath(), "path to projects.yaml")
	addr := flag.String("addr", "127.0.0.1:1325", "listen address (localhost only)")
	allowNonLocalhost := flag.Bool("allow-non-localhost", false, "allow bind on non-loopback")
	flag.Parse()

	if !*allowNonLocalhost && !strings.HasPrefix(*addr, "127.0.0.1:") && !strings.HasPrefix(*addr, "localhost:") {
		log.Fatal("gitboard: refuse to bind outside localhost (use -allow-non-localhost to override)")
	}

	doc, err := config.Load(*projectsPath)
	if err != nil {
		log.Fatal(err)
	}

	run := cliexec.New()
	dash := dashboard.New(forge.NewGitHub(run), forge.NewGitLab(run))
	handler := server.NewMux(server.Options{
		Addr:     *addr,
		Projects: doc.Projects,
		Dash:     dash,
		Triage:   triage.NewFromEnv(),
	})

	log.Printf("gitboard: %s (%d projects, projects file %s)", *addr, len(doc.Projects), *projectsPath)
	log.Printf("gitboard: uses gh and glab CLIs; AI triage optional via GITBOARD_LLM_BASE_URL or POLYPUS_BASE_URL")
	if err := http.ListenAndServe(*addr, handler); err != nil {
		log.Fatal(err)
	}
}
