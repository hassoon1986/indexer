package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/algorand/go-algorand-sdk/encoding/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/algorand/indexer/api"
	"github.com/algorand/indexer/config"
	"github.com/algorand/indexer/fetcher"
	"github.com/algorand/indexer/idb"
	"github.com/algorand/indexer/importer"
	"github.com/algorand/indexer/types"
)

var (
	algodDataDir     string
	algodAddr        string
	algodToken       string
	daemonServerAddr string
	noAlgod          bool
	developerMode    bool
	tokenString      string
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "run indexer daemon",
	Long:  "run indexer daemon. Serve api on HTTP.",
	//Args:
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		config.BindFlags(cmd)
		err = configureLogger()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to configure logger: %v", err)
			os.Exit(1)
		}

		if algodDataDir == "" {
			algodDataDir = os.Getenv("ALGORAND_DATA")
		}
		opts := idb.IndexerDbOptions{}
		if noAlgod {
			opts.ReadOnly = true
		}
		db := globalIndexerDb(&opts)

		ctx, cf := context.WithCancel(context.Background())
		defer cf()
		var bot fetcher.Fetcher
		if noAlgod {
			logger.Info("algod block following disabled")
		} else if algodAddr != "" && algodToken != "" {
			bot, err = fetcher.ForNetAndToken(algodAddr, algodToken, logger)
			maybeFail(err, "fetcher setup, %v", err)
		} else if algodDataDir != "" {
			if genesisJSONPath == "" {
				genesisJSONPath = filepath.Join(algodDataDir, "genesis.json")
			}
			bot, err = fetcher.ForDataDir(algodDataDir, logger)
			maybeFail(err, "fetcher setup, %v", err)
		} else {
			// no algod was found
			noAlgod = true
		}
		if !noAlgod {
			// Only do this if we're going to be writing
			// to the db, to allow for read-only query
			// servers that hit the db backend.
			err := importer.ImportProto(db)
			maybeFail(err, "import proto, %v", err)
		}
		if bot != nil {
			maxRound, err := db.GetMaxRound()
			if err == nil {
				bot.SetNextRound(maxRound + 1)
			}
			bih := blockImporterHandler{
				imp:   importer.NewDBImporter(db),
				db:    db,
				round: maxRound,
			}
			bot.AddBlockHandler(&bih)
			bot.SetContext(ctx)
			go func() {
				bot.Run()
				cf()
			}()
		}

		tokenArray := make([]string, 0)
		if tokenString != "" {
			tokenArray = append(tokenArray, tokenString)
		}

		// TODO: trap SIGTERM and call cf() to exit gracefully
		logger.Infof("serving on %s", daemonServerAddr)
		api.Serve(ctx, daemonServerAddr, db, logger, tokenArray, developerMode)
	},
}

func init() {
	daemonCmd.Flags().StringVarP(&algodDataDir, "algod", "d", "", "path to algod data dir, or $ALGORAND_DATA")
	daemonCmd.Flags().StringVarP(&algodAddr, "algod-net", "", "", "host:port of algod")
	daemonCmd.Flags().StringVarP(&algodToken, "algod-token", "", "", "api access token for algod")
	daemonCmd.Flags().StringVarP(&genesisJSONPath, "genesis", "g", "", "path to genesis.json (defaults to genesis.json in algod data dir if that was set)")
	daemonCmd.Flags().StringVarP(&daemonServerAddr, "server", "S", ":8980", "host:port to serve API on (default :8980)")
	daemonCmd.Flags().BoolVarP(&noAlgod, "no-algod", "", false, "disable connecting to algod for block following")
	daemonCmd.Flags().StringVarP(&tokenString, "token", "t", "", "an optional auth token, when set REST calls must use this token in a bearer format, or in a 'X-Indexer-API-Token' header")
	daemonCmd.Flags().BoolVarP(&developerMode, "dev-mode", "", false, "allow performance intensive operations like searching for accounts at a particular round")

	viper.RegisterAlias("algod", "algod-data-dir")
	viper.RegisterAlias("algod-net", "algod-address")
	viper.RegisterAlias("server", "server-address")
	viper.RegisterAlias("token", "api-token")
}

type blockImporterHandler struct {
	imp   importer.Importer
	db    idb.IndexerDb
	round uint64
}

func (bih *blockImporterHandler) HandleBlock(block *types.EncodedBlockCert) {
	start := time.Now()
	if uint64(block.Block.Round) != bih.round+1 {
		logger.Errorf("received block %d when expecting %d", block.Block.Round, bih.round+1)
	}
	bih.imp.ImportDecodedBlock(block)
	importer.UpdateAccounting(bih.db, genesisJSONPath, logger)
	dt := time.Now().Sub(start)
	if len(block.Block.Payset) == 0 {
		// accounting won't have updated the round state, so we do it here
		stateJSONStr, err := db.GetMetastate("state")
		var state importer.ImportState
		if err == nil && stateJSONStr != "" {
			state, err = importer.ParseImportState(stateJSONStr)
			if err != nil {
				logger.WithError(err).Errorf("error parsing import state")
				panic("error parsing import state in bih")
			}
		}
		state.AccountRound = int64(block.Block.Round)
		stateJSONStr = string(json.Encode(state))
		err = db.SetMetastate("state", stateJSONStr)
		if err != nil {
			logger.WithError(err).Errorf("failed to save import state")
		}
	}
	logger.Infof("round r=%d (%d txn) imported in %s", block.Block.Round, len(block.Block.Payset), dt.String())
	bih.round = uint64(block.Block.Round)
}
