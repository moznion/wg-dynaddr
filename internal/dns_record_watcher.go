package internal

import (
	"net"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
)

// DNSRecordWatcher the methods for watcher of a DNS record.
type DNSRecordWatcher interface {
	// StartWatching starts DNS watching.
	StartWatching() error
	// WaitAddressesChanges waits for the address information has changed, and this returned channel provides the changed value.
	WaitAddressesChanges() <-chan []string
}

// DefaultDNSRecordWatcher is a default implementation of DNSRecordWatcher.
type DefaultDNSRecordWatcher struct {
	lookupInterval time.Duration
	hostName       string
	changedValueCh chan []string
	addrsSnapShot  map[string]struct{}
}

// NewDefaultDNSRecordWatcher creates an instance of the DefaultDNSRecordWatcher.
func NewDefaultDNSRecordWatcher(lookupInterval time.Duration, hostName string) *DefaultDNSRecordWatcher {
	return &DefaultDNSRecordWatcher{
		lookupInterval: lookupInterval,
		hostName:       hostName,
		changedValueCh: make(chan []string, 1),
		addrsSnapShot:  make(map[string]struct{}),
	}
}

// See DNSRecordWatcher.StartWatching.
func (w *DefaultDNSRecordWatcher) StartWatching() error {
	exch := make(chan struct{}, 1)

	ticker := time.Tick(w.lookupInterval)
	for range ticker {
		func() {
			log.Debug().Msg("start checking the address changes")
			select {
			case exch <- struct{}{}:
				defer func() {
					select {
					case <-exch:
					default:
					}
					log.Debug().Msg("finished checking the address changes")
				}()
			default:
				// don't allow the multiple execution, skip it
				log.Debug().Msg("another checking sequence is undergoing. it skips this")
				return
			}
			w.checkAddressChanges()
		}()
	}

	return nil
}

// See DNSRecordWatcher.WaitAddressesChanges
func (w *DefaultDNSRecordWatcher) WaitAddressesChanges() <-chan []string {
	return w.changedValueCh
}

func (w *DefaultDNSRecordWatcher) checkAddressChanges() {
	addrs, err := net.LookupHost(w.hostName)
	if err != nil {
		log.Error().Err(err).Msg("failed to lookup host; continue")
		return
	}

	currentAddrsSet := make(map[string]struct{})
	for _, addr := range addrs {
		currentAddrsSet[addr] = struct{}{}
	}

	if !reflect.DeepEqual(w.addrsSnapShot, currentAddrsSet) {
		w.notifyChangedAddresses(currentAddrsSet)
	}
	w.addrsSnapShot = currentAddrsSet
}

func (w *DefaultDNSRecordWatcher) notifyChangedAddresses(currentAddrsSet map[string]struct{}) {
	distinctAddrs := make([]string, len(currentAddrsSet))
	i := 0
	for addr := range currentAddrsSet {
		distinctAddrs[i] = addr
		i++
	}

	w.publishAddrs(distinctAddrs)
	log.Info().Str("hostName", w.hostName).Strs("addrs", distinctAddrs).Msg("addresses has changed; it published the new addrs")
}

func (w *DefaultDNSRecordWatcher) publishAddrs(currentAddrs []string) {
	select {
	case w.changedValueCh <- currentAddrs:
	default:
		select {
		case <-w.changedValueCh:
		default:
		}

		select {
		case w.changedValueCh <- currentAddrs:
		default:
		}
	}
}
