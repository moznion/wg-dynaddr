package main

import (
	"time"

	"github.com/moznion/wg-dynaddr/internal"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("launched")

	reloader, err := internal.NewDefaultNetworkInterfaceAddressReloader("wg0" /* TODO opt */)
	if err != nil {
		log.Fatal().Err(err).Msg("reloader creation error; it exits unexpectedly")
	}

	watcher := internal.NewDefaultDNSRecordWatcher(
		60*time.Second, // TODO opt
		"example.com",  // TODO opt
	)
	go func() {
		err := watcher.StartWatching()
		log.Fatal().Err(err).Msg("DNS watching error; it exits unexpectedly")
	}()

	for {
		addrs := <-watcher.WaitAddressesChanges()

		err := reloader.Reload(addrs)
		if err != nil {
			log.Fatal().Err(err).Msg("interface reloading error; it exits unexpectedly")
		}
	}
}
