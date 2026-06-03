package main

/*
#cgo CFLAGS: -I../../
#cgo linux LDFLAGS: -L ../../ -lntgcalls -lm -lz
#cgo darwin LDFLAGS: -L ../../ -lntgcalls -lc++ -lz -lbz2 -liconv -framework AVFoundation -framework AudioToolbox -framework CoreAudio -framework QuartzCore -framework CoreMedia -framework VideoToolbox -framework AppKit -framework Metal -framework MetalKit -framework OpenGL -framework IOSurface -framework ScreenCaptureKit

// Currently is supported only dynamically linked library on Windows due to
// https://github.com/golang/go/issues/63903
#cgo windows LDFLAGS: -L../../ -lntgcalls
#include "ntgcalls/ntgcalls.h"
#include "glibc_compatibility.h"
*/
import "C"

import (
	"context"
	"errors"
	"net/http"
	"net/http/pprof"
	"os"
	"time"

	"github.com/Laky-64/gologging"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/modules"
	"main/internal/platforms"
)

func main() {
	initLogger()
	defer config.CloseLogging()

	shutdownPlatforms, err := platforms.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize platforms: " + err.Error())
	}
	defer shutdownPlatforms()

	checkFFmpegAndFFprobe()

	if err := refreshDirs(); err != nil {
		gologging.Fatal("Failed to refresh directories: " + err.Error())
	}

	gologging.Debug("Initializing MongoDB...")

	closeDB, err := database.Init(config.MongoURI)
	if err != nil {
		gologging.Fatal("Failed to initialize database: " + err.Error())
	}
	defer closeDB()

	gologging.Info("Database connected successfully")

	if err := locales.Load(); err != nil {
		gologging.Fatal("Failed to load locales: " + err.Error())
	}

	gologging.Debug("Initializing clients...")

	shutdownCore, err := core.Init()
	if err != nil {
		gologging.Fatal("Failed to initialize core: " + err.Error())
	}
	defer shutdownCore()

	core.GetAssistantIndexFunc = database.AssistantIndex
	core.F = modules.F

	if err := database.RebalanceAssistantIndexes(core.Assistants.Count()); err != nil {
		gologging.Fatal("Failed to rebalance Assistants: " + err.Error())
	}

	modules.Init(core.Bot, core.Assistants)

	srv := startHTTPServer()
	defer func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			gologging.Error("HTTP server shutdown error: " + err.Error())
		}
	}()

	core.Bot.Idle()
}

func startHTTPServer() *http.Server {
	// pprof num mux dedicado e bind em loopback: antes era DefaultServeMux +
	// 0.0.0.0, expondo /debug/pprof/* publicamente (info-leak de memória/stacks
	// e DoS via profiling). Pra acessar de fora, use port-forward/túnel.
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	srv := &http.Server{
		Addr:              "127.0.0.1:" + config.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			gologging.Error("HTTP server error: " + err.Error())
		}
	}()

	return srv
}

func initLogger() {
	gologging.SetLevel(gologging.DebugLevel)
	gologging.SetOutput(config.LogWriter)

	l := gologging.GetLogger("ntgcalls")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	l = gologging.GetLogger("webrtc")
	l.SetLevel(gologging.ErrorLevel)
	l.SetOutput(config.LogWriter)

	gologging.GetLogger("Database").SetOutput(config.LogWriter)
}

func refreshDirs() error {
	dirs := []string{
		"./cache",
		"./downloads",
	}

	for _, dir := range dirs {

		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
