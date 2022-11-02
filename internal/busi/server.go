package busi

import (
	"context"
	"data-extraction-notify/pkg/models/busi"
	"data-extraction-notify/pkg/utils"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	Ctx context.Context
	Cf  utils.TomlConfig
	wg  sync.WaitGroup
}

func NewServer(ctx context.Context) *Server {
	return &Server{Ctx: ctx}
}

func (s *Server) initconfig() {
	if err := utils.InitConfFile(Flags.Config, &s.Cf); err != nil {
		log.Fatalf("Load configuration file err: %v", err)
	}

	// utils.EngineGroup = utils.NewEngineGroup(ctx, &[]utils.EngineInfo{{utils.DBANALYSIS, cf.DataInfra.DBobservable, nil}, {utils.DBSPRD, cf.Sprd.DBsprd, sprd.Tables}})
	utils.EngineGroup = utils.NewEngineGroup(s.Ctx, &[]utils.EngineInfo{{utils.DBExtract, s.Cf.DataExtraction.DB, busi.Tables}})

	utils.InitKVEngine(s.Ctx, s.Cf.DataExtraction.MQ, "", 0)
}

func (s *Server) setLogTimeformat() {
	timeFormater := new(log.TextFormatter)
	timeFormater.FullTimestamp = true
	logrus.SetFormatter(timeFormater)
}

func (s *Server) Start() {
	s.initconfig()
	s.setLogTimeformat()

	go HttpServerStart(s.Cf.DataExtraction.Addr)

	{
		s.wg.Add(2)
		ctx, cancel := context.WithCancel(s.Ctx)
		go NotifyServerStart(ctx, s.wg.Done, s.Cf.DataExtraction.Lotus0, s.Cf.DataExtraction.MQ)
		<-s.sigHandle()
		cancel()
		s.wg.Wait()
	}
}

func (s *Server) sigHandle() <-chan os.Signal {
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGTERM, syscall.SIGINT)

	return sigChannel
}
