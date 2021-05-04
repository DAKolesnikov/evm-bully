package command

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aurora-is-near/evm-bully/nearapi"
	"github.com/aurora-is-near/evm-bully/replayer"
)

// Replay implements the 'replay' command.
func Replay(argv0 string, args ...string) error {
	var (
		nodeURL      nodeURLFlag
		testnetFlags testnetFlags
	)
	fs := flag.NewFlagSet(argv0, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <evmContract>\n", argv0)
		fmt.Fprintf(os.Stderr, "Replay transactions to NEAR EVM.\n")
		fs.PrintDefaults()
	}
	accountID := fs.String("accountId", "", "Unique identifier for the account that will be used to sign this call")
	block := fs.Uint64("block", defaultBlockHeight, "Block height")
	dataDir := fs.String("datadir", defaultDataDir, "Data directory containing the database to read")
	defrost := fs.Bool("defrost", false, "Defrost the database first")
	gas := fs.Uint64("gas", defaultGas, "Max amount of gas this call can use (in gas units)")
	hash := fs.String("hash", defaultBlockhash, "Block hash")
	cfg := nearapi.GetConfig()
	nodeURL.registerFlag(fs, cfg)
	testnetFlags.registerFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *accountID == "" {
		return errors.New("option -accountId is mandatory")
	}
	chainID, testnet, err := testnetFlags.determineTestnet()
	if err != nil {
		return err
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return flag.ErrHelp
	}
	evmContract := fs.Arg(0)
	// load account
	c := nearapi.NewConnection(string(nodeURL))
	a, err := nearapi.LoadAccount(c, cfg, *accountID)
	if err != nil {
		return err
	}
	// run replayer
	r := replayer.Replayer{
		ChainID:     chainID,
		Gas:         *gas,
		DataDir:     *dataDir,
		Testnet:     testnet,
		BlockHeight: *block,
		BlockHash:   *hash,
		Defrost:     *defrost,
	}
	return r.Replay(a, evmContract)
}
