/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/nft-rainbow/dodoBot/service"
	"github.com/nft-rainbow/dodoBot/utils"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "deploy erc721 contract",
	Long: `The admin of the bot can use this cmd to deploy his own contract.`,
	Example: `botCMD deploy [name] [symbol] [type] [appAddress]
- name The name of the contract
- symbol The symbol of the NFT
- type The type of the contract including erc721 and erc1155
- appAddress The address of the NFTRainbow app`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		name, symbol, address, contractType := args[0], args[1], args[3], args[2]
		_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, address)
		if err != nil {
			fmt.Println(err)
			return
		}

		token, err := service.Login()
		if err != nil {
			fmt.Println(err)
			return
		}

		contractAddress, err := service.DeployContract(token, name, symbol, address, contractType)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(contractAddress)
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
