package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

func DaemonMain() {
	log, err := logger.NewLog(logger.All, logger.All, "")

	if err != nil {
		panic(err)
	}

	defer log.Close()
	opts, err := server.LoadConfigFromPath("/etc/webby/config.json")

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using default configuration due to errors")
	}

	if opts.Log != "" {
		log.LogInfo("Opening '" + opts.Log + "' for recording logs")
		err = log.OpenFile(opts.Log)

		if err != nil {
			log.LogErr("Could not open '" + opts.Log + "' for logging")
		}
	}

	printing, err := logger.LevelFromString(opts.LogLevelPrint)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' due to errors for printing")
	}

	recording, err := logger.LevelFromString(opts.LogLevelRecord)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' due to errors for recording")
	}

	log.Printing = printing
	log.Saving = recording

	srv, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		return
	}

	commandListener, err := NewDaemonListener(map[DaemonCommand]func(DaemonCommandArg) error{
		Restart: func(_ DaemonCommandArg) error {
			// When the `Server.Start()` function returns it is automatically called
			// again in a loop.
			return srv.Stop()
		},

		LogRecord: func(arg DaemonCommandArg) error {
			logLevel := logger.LogLevel(arg)
			logLevel, err := logger.CheckLogLevel(uint8(logLevel))

			if err != nil {
				log.LogWarn("invalid log level given, using 'All'")
			}

			log.Saving = logLevel
			return nil
		},

		LogPrint: func(arg DaemonCommandArg) error {
			logLevel := logger.LogLevel(arg)
			logLevel, err := logger.CheckLogLevel(uint8(logLevel))

			if err != nil {
				log.LogWarn("invalid log level given, using 'All'")
			}

			log.Printing = logLevel
			return nil
		},
	}, log)

	usingDaemonSocket := true

	if err != nil {
		log.LogErr(err.Error())
		log.LogErr("Could not open Unix Domain Socket")
		usingDaemonSocket = false
	} else {
		go commandListener.Listen()
	}

	sigtermChannel := make(chan os.Signal, 1)
	signal.Notify(sigtermChannel, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			// Will restart the server on close.
			log.LogErr(srv.Start().Error())
			srv, err = server.NewServer(opts, &log)

			if err != nil {
				log.LogErr(err.Error())
				return
			}
		}
	}()

	sig := <-sigtermChannel
	log.LogWarn("Received signal: " + sig.String())

	if usingDaemonSocket {
		log.LogInfo("Closing Unix Domain Socket...")
		commandListener.Close()
	}

	log.LogInfo("Closing log...")
	log.Close()
}
