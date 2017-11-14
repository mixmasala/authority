package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/katzenpost/authority/config"
	"github.com/stretchr/testify/require"
)

func TestServerSchedule(t *testing.T) {
	require := require.New(t)
	datadir, err := ioutil.TempDir("", "katzenpost_datadir")
	require.NoError(err, "ioutil.TempDir")
	rawCfg := []byte(fmt.Sprintf(`
[Authority]
IdentityPrivateKey = "yw/v8jk219VSiJV0JMcuoZrvG1mLLut77qZtW/2819sMT4sBTGEVQDpGCb9vPI4qYKgwSz4oT5uZdlpTHG7tyw=="
Address = "127.0.0.1:8080"
BaseURL = "/"
DataDir = "%s"`, datadir))
	t.Logf("using config \n %s", rawCfg)
	cfg, err := config.Load(rawCfg)
	require.NoError(err, "config.Load")
	clock := clockwork.NewFakeClockAt(time.Now())
	ctx := context.TODO() // XXX
	_, err = New(cfg, ctx, clock)
	require.NoError(err, "server.New")
	clock.Advance(time.Hour * 3)
	clock.Advance(time.Hour * 3)
	clock.Advance(time.Hour * 3)

}
