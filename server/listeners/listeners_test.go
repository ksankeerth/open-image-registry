package listeners

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dummyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func resetListenerManager() {
	lm = nil
	once = sync.Once{}
}

func TestGetListenerManager(t *testing.T) {
	lm = GetListenerManager()
	assert.NotNil(t, lm)
}

func TestRegisterListener(t *testing.T) {
	resetListenerManager()
	GetListenerManager()
	assert.NotNil(t, lm)

	t.Run("Not allowed to start registry listener with invalid ports", func(t *testing.T) {
		err := lm.RegisterListener("test-reg", "test-reg", 200, dummyHandler, 0)
		require.NotNil(t, err)
		assert.Equal(t, "invalid regId or port", err.Error())
	})

	t.Run("Not allowed to start registry listener with invalid regId", func(t *testing.T) {
		err := lm.RegisterListener("", "test-reg", 2000, dummyHandler, 0)
		require.NotNil(t, err)
		assert.Equal(t, "invalid regId or port", err.Error())
	})

	t.Run("Not allowed to start registry listener with same port again", func(t *testing.T) {
		err := lm.RegisterListener("test-reg", "test-reg", 2000, dummyHandler, 0)
		require.Nil(t, err)

		err = lm.RegisterListener("test-reg1", "test-reg1", 2000, dummyHandler, 0)
		require.NotNil(t, err)
		assert.Equal(t, "port is not available", err.Error())
	})

	t.Run("Not allowed to start registry listener with same regId again", func(t *testing.T) {
		err := lm.RegisterListener("test-reg1", "test-reg1", 3000, dummyHandler, 0)
		require.Nil(t, err)

		err = lm.RegisterListener("test-reg1", "test-reg1", 3001, dummyHandler, 0)
		require.NotNil(t, err)
		assert.Equal(t, "listener already exists", err.Error())
	})

	t.Run("Release locks on errors", func(t *testing.T) {
		err := lm.RegisterListener("", "test-reg2", 4000, dummyHandler, 0)
		require.NotNil(t, err)
		assert.Equal(t, "invalid regId or port", err.Error())

		err = lm.RegisterListener("test-reg2", "test-reg2", 4000, dummyHandler, 0)
		require.Nil(t, err)
	})

	t.Run("Start Listener after the delay", func(t *testing.T) {
		err := lm.RegisterListener("test-reg3", "test-reg3", 5000, dummyHandler, 5)
		require.Nil(t, err)

		_, err = http.Get("http://localhost:5000/")
		require.NotNil(t, err)

		time.Sleep(5 * time.Second)
		res, err := http.Get("http://localhost:5000/")
		require.Nil(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestUnregisterListener(t *testing.T) {
	resetListenerManager()
	GetListenerManager()
	assert.NotNil(t, lm)

	t.Run("Unregister none existent listener won't cause error", func(t *testing.T) {
		err := lm.UnregisterListener("reg1", 0)
		require.Nil(t, err)
	})

	t.Run("Unregister timeout", func(t *testing.T) {
		err := lm.RegisterListener("reg2", "reg2", 8000, dummyHandler, 5)
		require.Nil(t, err)

		lm.mu.Lock()
		ln, ok := lm.listeners["reg2"]
		require.NotNil(t, ln)
		require.True(t, ok)
		// Replace with another done channel, so server won't close this new channel
		ln.Done = make(chan struct{})
		lm.mu.Unlock()

		err = lm.UnregisterListener("reg2", 5)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "Timeout occurred while shutting down listener for registry")

		lm.mu.Lock()
		ln, ok = lm.listeners["reg2"]
		require.Nil(t, ln)
		require.False(t, ok)
		lm.mu.Unlock()
	})
}
