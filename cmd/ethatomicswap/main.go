// Copyright (c) 2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	rpc "github.com/ethereum/go-ethereum/rpc"
)

const verify = true

const secretSize = 32

// Type Block
type Block struct {
	Number string
	Hash   string
}

var (
	flagset     = flag.NewFlagSet("", flag.ExitOnError)
	connectFlag = flagset.String("s", "localhost", "host[:port] of Geth RPC server")
	rpcuserFlag = flagset.String("rpcuser", "", "username for wallet RPC authentication")
	rpcpassFlag = flagset.String("rpcpass", "", "password for wallet RPC authentication")
	testnetFlag = flagset.Bool("testnet", false, "use testnet ropsten network")
)

// There are two directions that the atomic swap can be performed, as the
// initiator can be on either chain.  This tool only deals with creating the
// Bitcoin transactions for these swaps.  A second tool should be used for the
// transaction on the other chain.  Any chain can be used so long as it supports
// OP_RIPEMD160 and OP_CHECKLOCKTIMEVERIFY.
//
// Example scenerios using bitcoin as the second chain:
//
// Scenerio 1:
//   cp1 initiates (dcr)
//   cp2 participates with cp1 H(S) (btc)
//   cp1 redeems btc revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems dcr with S
//
// Scenerio 2:
//   cp1 initiates (btc)
//   cp2 participates with cp1 H(S) (dcr)
//   cp1 redeems dcr revealing S
//     - must verify H(S) in contract is hash of known secret
//   cp2 redeems btc with S

func init() {
	flagset.Usage = func() {
		fmt.Println("Usage: ethatomicswap [flags] cmd [cmd args]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  initiate <participant address> <amount>")
		fmt.Println("  participate <initiator address> <amount> <secret hash>")
		fmt.Println("  redeem <contract> <contract transaction> <secret>")
		fmt.Println("  extractsecret <redemption transaction> <secret hash>")
		fmt.Println("  auditcontract <contract> <contract transaction>")
		fmt.Println()
		fmt.Println("Flags:")
		flagset.PrintDefaults()
	}
}

type command interface {
	runCommand(context.Context, *rpc.Client) error
}

type initiateCmd struct {
}

type participateCmd struct {
	cp1Addr    common.Address
	amount     *big.Int
	secretHash []byte
}

type redeemCmd struct {
	contract   []byte
	contractTx *types.Transaction
	secret     []byte
}

type extractSecretCmd struct {
	redemptionTx *types.Transaction
	secretHash   []byte
}

type auditContractCmd struct {
	contract   []byte
	contractTx *types.Transaction
}

func main() {
	err, showUsage := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if showUsage {
		flagset.Usage()
	}
	if err != nil || showUsage {
		os.Exit(1)
	}
}

func run() (err error, showUsage bool) {
	flagset.Parse(os.Args[1:])
	args := flagset.Args()
	if len(args) == 0 {
		return nil, true
	}
	cmdArgs := 0
	switch args[0] {
	case "initiate":
		cmdArgs = 2
	case "participate":
		cmdArgs = 3
	case "redeem":
		cmdArgs = 3
	case "extractsecret":
		cmdArgs = 2
	case "auditcontract":
		cmdArgs = 2
	default:
		return fmt.Errorf("unknown command %v", args[0]), true
	}
	nArgs := checkCmdArgLength(args[1:], cmdArgs)
	flagset.Parse(args[1+nArgs:])
	if nArgs < cmdArgs {
		return fmt.Errorf("%s: too few arguments", args[0]), true
	}
	if flagset.NArg() != 0 {
		return fmt.Errorf("unexpected argument: %s", flagset.Arg(0)), true
	}

	var cmd command
	switch args[0] {
	case "initiate":
		cmd = &initiateCmd{}

	case "participate":

	case "redeem":

	case "extractsecret":

	case "auditcontract":

	}

	connect, err := normalizeAddress(*connectFlag, walletPort("mainnet"))
	if err != nil {
		return fmt.Errorf("geth server address: %v", err), true
	}

	client, err := rpc.Dial(connect)
	if err != nil {
		return fmt.Errorf("rpc connect: %v", err), false
	}
	defer func() {
		client.Close()
	}()

	err = cmd.runCommand(context.Background(), client)
	return err, false
}

func checkCmdArgLength(args []string, required int) (nArgs int) {
	if len(args) < required {
		return 0
	}
	for i, arg := range args[:required] {
		if len(arg) != 1 && strings.HasPrefix(arg, "-") {
			return i
		}
	}
	return required
}

func (cmd *initiateCmd) runCommand(ctx context.Context, c *rpc.Client) error {
	var lastBlock Block
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := c.CallContext(ctx, &lastBlock, "eth_getBlockByNumber", "latest", false); err != nil {
		println("can't get latest block:", err)
		return err
	}
	fmt.Println(lastBlock.Number)
	fmt.Println(lastBlock.Hash)
	return nil
}

func normalizeAddress(addr string, defaultPort string) (hostport string, err error) {
	host, port, origErr := net.SplitHostPort(addr)
	if origErr == nil {
		return net.JoinHostPort(host, port), nil
	}
	addr = net.JoinHostPort(addr, defaultPort)
	_, _, err = net.SplitHostPort(addr)
	if err != nil {
		return "", origErr
	}
	return "http://" + addr, nil
}

func walletPort(params string) string {
	switch params {
	case "testnet":
		return "8545"
	case "mainnet":
		return "8545"
	default:
		return "8545"
	}
}
