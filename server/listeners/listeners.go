package listeners

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ksankeerth/open-image-registry/lib"
	"github.com/ksankeerth/open-image-registry/log"
)

var once sync.Once

var lm *ListenerManager

// ListenerManager manages the lifecycle of HTTP listeners.
//
// Typical use cases:
//  1. Start listeners during system startup for the hosted registry and upstream proxies.
//  2. Register a new listener when a new upstream proxy is created.
//  3. Unregister (stop) a listener when an upstream proxy is disabled.
//  4. Re-register (start) a listener when an upstream proxy is re-enabled.
type ListenerManager struct {
	portsLock        *lib.KeyLock
	regListenersLock *lib.KeyLock
	listeners        map[string]*RegistryListener
	mu               sync.Mutex
}

type RegistryListener struct {
	RegId   string
	RegName string
	Server  *http.Server
	Cancel  context.CancelFunc
	Done    chan struct{}
	Port    uint
}

func GetListenerManager() *ListenerManager {
	once.Do(func() {
		lm = &ListenerManager{
			listeners:        make(map[string]*RegistryListener),
			portsLock:        lib.NewKeyLock(),
			regListenersLock: lib.NewKeyLock(),
		}
	})
	return lm
}

// RegisterListener will registery new listner for given values. Creating new listner
// and starting to serve requests will happen asynchronolously.
// TODO later we could imporove each listeners to have different timeouts. For now,
// We'll keep this simple for now.
func (lm *ListenerManager) RegisterListener(regId, regName string, port uint, h http.Handler,
	listenDelayInSeconds time.Duration) error {
	if regId == "" || port < 1025 {
		log.Logger().Warn().Msgf("Invalid regId or port.")
		return fmt.Errorf("invalid regId or port")
	}

	locked := lm.portsLock.Lock(strconv.Itoa(int(port)))
	if !locked {
		log.Logger().Warn().Msgf("Port(%d) is already taken by another registry.", port)
		return fmt.Errorf("port is not available")
	}

	locked = lm.regListenersLock.Lock(regId)
	if !locked {
		log.Logger().Warn().Msgf("Registry(%s) has listener already", regId)
		lm.portsLock.Unlock(strconv.Itoa(int(port)))
		return fmt.Errorf("listener already exists")
	}

	lm.mu.Lock()
	if ln, ok := lm.listeners[regId]; ok {
		log.Logger().Error().Msgf("Existing HTTP listener with address: %s found for registry: %s", ln.Server.Addr, regId)
		lm.mu.Unlock()
		lm.cleanup(regId, port)
		return fmt.Errorf("Registry(%s) has a HTTP listener already.", regId)
	}
	lm.mu.Unlock()

	addr := fmt.Sprintf(":%d", port)
	// for _, ln := range lm.listeners {
	// 	if ln.Server.Addr == addr {
	// 		log.Logger().Error().Msgf("Address: %s is already occupied by registry: %s", addr, ln.RegId)
	// 		return fmt.Errorf("Address: %s is already occupied by registry: %s", addr, ln.RegId)
	// 	}
	// }

	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:    addr,
		Handler: h,
	}

	regLn := &RegistryListener{
		RegId:   regId,
		RegName: regName,
		Server:  server,
		Cancel:  cancel,
		Done:    make(chan struct{}),
		Port:    port,
	}

	lm.mu.Lock()
	lm.listeners[regId] = regLn
	lm.mu.Unlock()

	go func() {
		defer close(regLn.Done)
		time.Sleep(listenDelayInSeconds * time.Second)
		log.Logger().Info().Msgf("Listener is about to start on port %d for registry %s", port, regName)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger().Error().Err(err).Msgf("Registry Listener went shutdown due to errors: %s", regId)

			lm.cleanup(regId, port)
		}
	}()

	go func() {
		<-ctx.Done()
		log.Logger().Info().Msgf("Shutting down HTTP listener for registry: %s", regId)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Logger().Error().Err(err).Msgf("Error occurred while shutting down HTTP listener for registry: %s", regId)
			server.Close()
		}

		lm.cleanup(regId, port)

	}()

	return nil
}

func (lm *ListenerManager) UnregisterListener(regId string, waitTimeout time.Duration) error {
	regLn, ok := lm.listeners[regId]
	if !ok {
		return nil
	}
	// trigger shutdown
	regLn.Cancel()

	// Wait for server shutdown
	select {
	case <-regLn.Done:
		return nil
	case <-time.After(waitTimeout * time.Second):
		lm.cleanup(regId, regLn.Port)
		return fmt.Errorf("Timeout occurred while shutting down listener for registry: %s", regId)
	}
}

func (lm *ListenerManager) cleanup(regId string, port uint) {
	lm.mu.Lock()
	delete(lm.listeners, regId)
	lm.mu.Unlock()
	lm.portsLock.Unlock(strconv.Itoa(int(port)))
	lm.regListenersLock.Unlock(regId)
}
