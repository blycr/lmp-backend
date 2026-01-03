package main

import (
	"fmt"
	"log"
	"net/http"

	"lms/backend/internal/api"
	"lms/backend/internal/config"
	"lms/backend/internal/files"
	"lms/backend/internal/auth"
)

func main() {
	mgr, err := config.NewManager()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	provider := func() *config.Config { return mgr.Current() }
	am := auth.NewManager(provider().Auth.SessionTimeout)
	auth.SetDefaultManager(am)
	ix := files.NewIndexer()
	ix.SetRoots(provider().Files.ShareDirs)
	_ = ix.Rebuild()
	_ = ix.StartWatching()
	files.SetDefaultIndexer(ix)
	mgr.Subscribe(func(c *config.Config) {
		am = auth.NewManager(c.Auth.SessionTimeout)
		auth.SetDefaultManager(am)
		ix.SetRoots(c.Files.ShareDirs)
		_ = ix.Rebuild()
		_ = ix.StartWatching()
	})
	r := api.SetupRouter(provider)

	addr := fmt.Sprintf(":%d", provider().Server.Port)
	log.Printf("LMP backend listening on %s (lan_only=%v)", addr, provider().Server.LANOnly)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
