package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	dodo "github.com/dodo-open/dodo-open-go"
	"github.com/dodo-open/dodo-open-go/client"
	"github.com/dodo-open/dodo-open-go/model"
	"github.com/dodo-open/dodo-open-go/tools"
	"github.com/dodo-open/dodo-open-go/websocket"
	"github.com/nft-rainbow/dodoBot/database"
	"github.com/nft-rainbow/dodoBot/models"
	"github.com/nft-rainbow/dodoBot/service"
	"github.com/nft-rainbow/dodoBot/utils"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"
)

func initConfig() {
	viper.SetConfigName("config")             // name of config file (without extension)
	viper.SetConfigType("yaml")               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")                  // optionally look for config in the working directory
	err := viper.ReadInConfig()               // Find and read the config file
	if err != nil {                           // Handle errors reading the config file
		log.Fatalln(fmt.Errorf("fatal error config file: %w", err))
	}
}

func init() {
	initConfig()
}

func main() {
	database.ConnectDB()
	instance, err := dodo.NewInstance(viper.GetString("bot.clientId"), viper.GetString("bot.tokenId"), client.WithTimeout(time.Second*3))
	if err != nil {
		panic(err)
	}
	handlers := &websocket.MessageHandlers{
		ChannelMessage: func(event *websocket.WSEventMessage, data *websocket.ChannelMessageEventBody) error {
			switch data.MessageType {
			case model.TextMsg:
				messageBody := &model.TextMessage{}
				if err := tools.JSON.Unmarshal(data.MessageBody, &messageBody); err != nil {
					return err
				}

				if strings.Contains(messageBody.Content, "/claim easyMint") {
					tmp := strings.Split(messageBody.Content, " ")
					if len(tmp) < 3 {
						_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
							ChannelId: data.ChannelId,
							MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> %s", data.DodoId, "The input is wrong")},
						})
						return nil
					}
					userAddress := tmp[2]
					_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
					if err != nil {
						_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
							ChannelId: data.ChannelId,
							MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> %s", data.DodoId, err.Error())},
						})
						return nil
					}
					err = checkRestrain(userAddress, database.EasyMintBucket)
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), "", nil)
						return nil
					}

					token, err := service.Login()
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.EasyMintBucket)
						return nil
					}

					resp, err := service.SendEasyMintRequest(token, models.EasyMintMetaDto{
						Chain:         viper.GetString("chainType"),
						Name:          viper.GetString("easyMint.name"),
						Description:   viper.GetString("easyMint.description"),
						MintToAddress: userAddress,
						FileUrl:       viper.GetString("easyMint.fileUrl"),
					})
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.EasyMintBucket)
						return nil
					}
					_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
						ChannelId: data.ChannelId,
						MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", data.DodoId, resp.UserAddress, resp.NFTAddress, resp.Advertise)},
						DodoId: data.DodoId,
					})
					return nil

				}

				if strings.Contains(messageBody.Content, "/claim customMint") {
					contents := strings.Split(messageBody.Content, " ")
					if len(contents) < 3 {
						_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
							ChannelId: data.ChannelId,
							MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> %s", data.DodoId, "The input is wrong")},
						})
						return nil
					}
					userAddress := contents[2]
					_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
					if err != nil {
						_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
							ChannelId: data.ChannelId,
							MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> %s", data.DodoId, err.Error())},
						})
						return nil
					}

					contractAddress := viper.GetString("customMint.contractAddress")
					_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, contractAddress)
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), "", database.CustomMintBucket)
						return nil
					}
					err = checkRestrain(userAddress, database.CustomMintBucket)
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.CustomMintBucket)
						return nil
					}

					token, err := service.Login()
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.CustomMintBucket)
						return nil
					}

					metadataUri, err := service.CreateMetadata(token, viper.GetString("customMint.fileUrl"), viper.GetString("customMint.name"), viper.GetString("customMint.description"))
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.CustomMintBucket)
						return nil
					}

					resp , err := service.SendCustomMintRequest(token, models.CustomMintDto{
						models.ContractInfoDto{
							Chain: viper.GetString("chainType"),
							ContractType: viper.GetString("customMint.contractType"),
							ContractAddress: contractAddress,
						},
						models.MintItemDto{
							MintToAddress: userAddress,
							MetadataUri: metadataUri,
						},
					})
					if err != nil {
						processErrorMessage(&instance, data, err.Error(), userAddress, database.CustomMintBucket)
						return nil
					}
					_, _ = instance.SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
						ChannelId: data.ChannelId,
						MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", data.DodoId, resp.UserAddress, resp.NFTAddress, resp.Advertise)},
						DodoId: data.DodoId,
					})
					return nil
				}
				}
			return nil
		},
	}

	ws, err := websocket.New(instance,
		websocket.WithMessageQueueSize(128),
		websocket.WithMessageHandlers(handlers),
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Start to connect")

	err = ws.Connect()
	if err != nil {
		panic(err)
	}
	fmt.Println("Start to listen")
	err = ws.Listen()
	if err != nil {
		panic(err)
	}
}

func checkRestrain(address string, mintType []byte) error{
	count, err := database.GetCount(address, mintType)
	if err != nil {
		return err
	}
	if count == nil {
		err = database.InsertDB(address, []byte("1"), mintType)
		if err != nil {
			return err
		}
		return nil
	}

	if !bytes.Equal(count, []byte("0")) {
		return errors.New("This address has minted the NFT")
	}

	return nil
}

func processErrorMessage(instance *client.Client, data *websocket.ChannelMessageEventBody, message, address string, mintType []byte) {
	_, _ = (*instance).SendChannelMessage(context.Background(), &model.SendChannelMessageReq{
		ChannelId: data.ChannelId,
		MessageBody: &model.TextMessage{Content: fmt.Sprintf("<@!%s> %s", data.DodoId, message)},
	})

	if address != "" {
		_ = database.InsertDB(address, []byte("0"), mintType)
	}
	return
}



