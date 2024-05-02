package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/eaglelabs-consulting/playground/framework"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
	envconfig "github.com/sethvargo/go-envconfig"
)

type DeploymentConfig struct {
	Environment string `env:"ENV, default=prod"`
}

func main() {
	var deploymentConfig DeploymentConfig
	var cfg framework.Config

	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		log.Fatal(err)
	}
	if err := envconfig.Process(context.Background(), &deploymentConfig); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SUAVE Signer Address: %s\n", cfg.FundedAccount.Address())
	fmt.Printf("L1 Signer Address: %s\n", cfg.FundedAccountL1.Address())
	fmt.Printf("L1 RPC: %s\n", cfg.L1RPC)
	fmt.Printf("Kettle RPC: %s\n", cfg.KettleRPC)

	var fr *framework.Framework
	if deploymentConfig.Environment == "prod" {
		fmt.Print("Deploying using SUAVE and L1\n")
		fr = framework.New(framework.WithL1())
		// ethClient := fr.L1.RPC()
	} else {
		fmt.Print("Deploying using local suave-geth\n")
		fr = framework.New()
	}
	suaveChainId, _ := fr.Suave.RPC().ChainID(context.Background())

	var chatgptBug *framework.Contract
	// fetch ChatGPT API key from .env file
	chatgptApiKey := os.Getenv("CHAT_GPT_KEY")

	fmt.Printf("Deploying ChatGPTBug with key %s\n", chatgptApiKey)

	chatgptBug = fr.Suave.DeployContract("ChatGPTBug.sol/ChatGPTBug.json")

	// register ChatGPT key using confidential request
	chatgptBug.SendConfidentialRequest("registerKeyOffchain", []interface{}{}, []byte(chatgptApiKey))

	auth, err := bind.NewKeyedTransactorWithChainID(cfg.FundedAccount.Priv, suaveChainId)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}

	offchain(cfg.FundedAccount, chatgptBug, fr.Suave.RPC(), auth)
}

func offchain(privKey *framework.PrivKey, contractInstance *framework.Contract, client *ethclient.Client, auth *bind.TransactOpts) {
	fmt.Print("Offchain call\n")

	chatgptbug := contractInstance.Ref(privKey)
	chatgptbug.SendConfidentialRequest("offchain", []interface{}{}, nil)
}
