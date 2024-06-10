package integration_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/database/memdb"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/choices"
	"github.com/ava-labs/avalanchego/snow/consensus/snowman"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/avalanchego/snow/validators"
	"github.com/ava-labs/avalanchego/utils/crypto/bls"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/utils/set"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/crypto/ed25519"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/actions"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/auth"
	lconsts "github.com/ava-labs/hypersdk/examples/typescriptvm/consts"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/controller"
	"github.com/ava-labs/hypersdk/examples/typescriptvm/genesis"
	lrpc "github.com/ava-labs/hypersdk/examples/typescriptvm/rpc"
	"github.com/ava-labs/hypersdk/fees"
	"github.com/ava-labs/hypersdk/rpc"
	"github.com/ava-labs/hypersdk/vm"
	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const VM_COUNT = 3

type instance struct {
	chainID           ids.ID
	nodeID            ids.NodeID
	vm                *vm.VM
	toEngine          chan common.Message
	JSONRPCServer     *httptest.Server
	BaseJSONRPCServer *httptest.Server
	WebSocketServer   *httptest.Server
	cli               *rpc.JSONRPCClient // clients for embedded VMs
	lcli              *lrpc.JSONRPCClient
}

func TestMain(m *testing.M) {
	//clean up before tests to avoid disk space issues
	removeAllGlob("./*.log")
	removeAllGlob("/tmp/NodeID-*-chainData*")

	// Run all tests
	code := m.Run()

	// Exit with the result of the test run
	os.Exit(code)
}

func (prep *prepeareResult) expectBlk(t *testing.T, i instance) func(bool) []*chain.Result {
	ctx := context.TODO()

	// manually signal ready
	require.NoError(t, i.vm.Builder().Force(ctx))
	// manually ack ready sig as in engine
	<-i.toEngine

	blk, err := i.vm.BuildBlock(ctx)
	if err != nil {
		panic(err)
	}
	require.NoError(t, err)
	require.NotNil(t, blk)

	require.NoError(t, blk.Verify(ctx))
	require.Equal(t, choices.Processing, blk.Status())

	err = i.vm.SetPreference(ctx, blk.ID())
	require.NoError(t, err)

	return func(add bool) []*chain.Result {
		require.NoError(t, blk.Accept(ctx))
		require.Equal(t, choices.Accepted, blk.Status())

		if add {
			prep.blocks = append(prep.blocks, blk)
		}

		lastAccepted, err := i.vm.LastAccepted(ctx)
		require.NoError(t, err)
		require.Equal(t, lastAccepted, blk.ID())
		return blk.(*chain.StatelessBlock).Results()
	}
}

type prepeareResult struct {
	instance instance
	priv     ed25519.PrivateKey
	pk       ed25519.PublicKey
	factory  *auth.ED25519Factory
	addr     codec.Address
	addrStr  string

	priv2    ed25519.PrivateKey
	pk2      ed25519.PublicKey
	factory2 *auth.ED25519Factory
	addr2    codec.Address
	addrStr2 string

	priv3    ed25519.PrivateKey
	pk3      ed25519.PublicKey
	factory3 *auth.ED25519Factory
	addr3    codec.Address
	addrStr3 string

	blocks []snowman.Block
}

func prepare(t *testing.T) prepeareResult {
	prep := prepeareResult{}

	var (
		// when used with embedded VMs
		genesisBytes []byte

		networkID  uint32
		gen        *genesis.Genesis
		logFactory logging.Factory
		log        logging.Logger
	)

	logFactory = logging.NewFactory(logging.Config{
		DisplayLevel: logging.Debug,
	})
	l, err := logFactory.Make("main")
	if err != nil {
		panic(err)
	}
	log = l

	log.Info("VMID", zap.Stringer("id", lconsts.ID))
	require.Greater(t, VM_COUNT, 1)

	prep.priv, err = ed25519.GeneratePrivateKey()
	require.NoError(t, err)

	prep.pk = prep.priv.PublicKey()
	prep.factory = auth.NewED25519Factory(prep.priv)
	prep.addr = auth.NewED25519Address(prep.pk)
	prep.addrStr = codec.MustAddressBech32(lconsts.HRP, prep.addr)
	log.Debug(
		"generated key",
		zap.String("addr", prep.addrStr),
		zap.String("pk", hex.EncodeToString(prep.priv[:])),
	)

	prep.priv2, err = ed25519.GeneratePrivateKey()
	require.NoError(t, err)

	prep.pk2 = prep.priv2.PublicKey()
	prep.factory2 = auth.NewED25519Factory(prep.priv2)
	prep.addr2 = auth.NewED25519Address(prep.pk2)
	prep.addrStr2 = codec.MustAddressBech32(lconsts.HRP, prep.addr2)
	log.Debug(
		"generated key",
		zap.String("addr", prep.addrStr2),
		zap.String("pk", hex.EncodeToString(prep.priv2[:])),
	)

	prep.priv3, err = ed25519.GeneratePrivateKey()
	require.NoError(t, err)

	prep.pk3 = prep.priv3.PublicKey()
	prep.factory3 = auth.NewED25519Factory(prep.priv3)
	prep.addr3 = auth.NewED25519Address(prep.pk3)
	prep.addrStr3 = codec.MustAddressBech32(lconsts.HRP, prep.addr3)
	log.Debug(
		"generated key",
		zap.String("addr", prep.addrStr3),
		zap.String("pk", hex.EncodeToString(prep.priv3[:])),
	)

	// create embedded VMs
	instances := make([]instance, VM_COUNT)

	gen = genesis.Default()
	gen.MinUnitPrice = fees.Dimensions{1, 1, 1, 1, 1}
	gen.MinBlockGap = 0
	gen.CustomAllocation = []*genesis.CustomAllocation{
		{
			Address: prep.addrStr,
			Balance: 10_000_000,
		},
	}
	genesisBytes, err = json.Marshal(gen)
	require.NoError(t, err)

	networkID = uint32(1)
	subnetID := ids.GenerateTestID()
	chainID := ids.GenerateTestID()

	app := &appSender{}
	for i := range instances {
		nodeID := ids.GenerateTestNodeID()
		sk, err := bls.NewSecretKey()
		require.NoError(t, err)
		l, err := logFactory.Make(nodeID.String())
		require.NoError(t, err)
		dname, err := os.MkdirTemp("", fmt.Sprintf("%s-chainData", nodeID.String()))
		require.NoError(t, err)
		snowCtx := &snow.Context{
			NetworkID:      networkID,
			SubnetID:       subnetID,
			ChainID:        chainID,
			NodeID:         nodeID,
			Log:            l,
			ChainDataDir:   dname,
			Metrics:        metrics.NewOptionalGatherer(),
			PublicKey:      bls.PublicFromSecretKey(sk),
			WarpSigner:     warp.NewSigner(sk, networkID, chainID),
			ValidatorState: &validators.TestState{},
		}

		toEngine := make(chan common.Message, 1)
		db := memdb.New()

		v := controller.New()
		err = v.Initialize(
			context.TODO(),
			snowCtx,
			db,
			genesisBytes,
			nil,
			[]byte(
				`{"parallelism":3, "testMode":true, "logLevel":"debug"}`,
			),
			toEngine,
			nil,
			app,
		)
		require.NoError(t, err)

		var hd map[string]http.Handler
		hd, err = v.CreateHandlers(context.TODO())
		require.NoError(t, err)

		jsonRPCServer := httptest.NewServer(hd[rpc.JSONRPCEndpoint])
		ljsonRPCServer := httptest.NewServer(hd[lrpc.JSONRPCEndpoint])
		webSocketServer := httptest.NewServer(hd[rpc.WebSocketEndpoint])
		instances[i] = instance{
			chainID:           snowCtx.ChainID,
			nodeID:            snowCtx.NodeID,
			vm:                v,
			toEngine:          toEngine,
			JSONRPCServer:     jsonRPCServer,
			BaseJSONRPCServer: ljsonRPCServer,
			WebSocketServer:   webSocketServer,
			cli:               rpc.NewJSONRPCClient(jsonRPCServer.URL),
			lcli:              lrpc.NewJSONRPCClient(ljsonRPCServer.URL, snowCtx.NetworkID, snowCtx.ChainID),
		}

		// Force sync ready (to mimic bootstrapping from genesis)
		v.ForceReady()
	}

	// Verify genesis allocates loaded correctly (do here otherwise test may
	// check during and it will be inaccurate)
	for i, inst := range instances {
		cli := inst.lcli
		g, err := cli.Genesis(context.Background())
		// require.NoError(t, err)
		require.NoError(t, err)

		csupply := uint64(0)
		for _, alloc := range g.CustomAllocation {
			balance, err := cli.Balance(context.Background(), alloc.Address)
			require.NoError(t, err, "failed to get balance from instance %d", i)
			require.Equal(t, alloc.Balance, balance)
			log.Warn("balances", zap.String("addr", alloc.Address), zap.Uint64("bal", balance))
			csupply += alloc.Balance
		}
	}
	prep.blocks = []snowman.Block{}

	app.instances = instances
	color.Blue("created %d VMs", VM_COUNT)

	prep.instance = instances[0]

	return prep
}

var _ common.AppSender = &appSender{}

type appSender struct {
	next      int
	instances []instance
}

func (app *appSender) SendAppGossip(ctx context.Context, _ common.SendConfig, appGossipBytes []byte) error {
	n := len(app.instances)
	sender := app.instances[app.next].nodeID
	app.next++
	app.next %= n
	return app.instances[app.next].vm.AppGossip(ctx, sender, appGossipBytes)
}

func (*appSender) SendAppRequest(context.Context, set.Set[ids.NodeID], uint32, []byte) error {
	return nil
}

func (*appSender) SendAppError(context.Context, ids.NodeID, uint32, int32, string) error {
	return nil
}

func (*appSender) SendAppResponse(context.Context, ids.NodeID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppRequest(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppResponse(context.Context, ids.ID, uint32, []byte) error {
	return nil
}

func (*appSender) SendCrossChainAppError(context.Context, ids.ID, uint32, int32, string) error {
	return nil
}

func removeAllGlob(pattern string) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("Error finding files with pattern %s: %v", pattern, err)
	}
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			log.Fatalf("Failed to delete %s: %v", file, err)
		}
	}
	log.Printf("All files and directories matching pattern %s have been deleted", pattern)
}

func deployTestContractHelper(t *testing.T, prep prepeareResult) string {
	discriminator := 111

	//send CreateContract tx

	parser, err := prep.instance.lcli.Parser(context.Background())
	require.NoError(t, err)
	submit, _, _, err := prep.instance.cli.GenerateTransaction(
		context.Background(),
		parser,
		[]chain.Action{&actions.CreateContract{
			Bytecode:      testWasmBytes,
			Discriminator: uint16(discriminator),
		}},
		prep.factory,
	)
	require.NoError(t, err)
	require.NoError(t, submit(context.Background()))

	results := prep.expectBlk(t, prep.instance)(false)
	require.Len(t, results, 1)
	require.True(t, results[0].Success)

	//contract deployed, get address from output
	contractAddrString := string(results[0].Outputs[0][0])

	return contractAddrString
}
