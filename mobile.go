package neoutils

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/tsfdsong/neo-utils/nep6"
	"github.com/tsfdsong/neo-utils/o3"
	"github.com/tsfdsong/neo-utils/smartcontract"
)

// This class contains simplified method designed specifically for gomobile bind
// gomobile bind doesn't support slice argument or return

func utxoFromO3Platform(network string, address string) (smartcontract.Unspent, error) {

	unspent := smartcontract.Unspent{
		Assets: map[smartcontract.NativeAsset]*smartcontract.Balance{},
	}

	client := o3.DefaultO3APIClient()
	if network == "test" {
		client = o3.APIClientWithNEOTestnet()
	}

	if network == "private" {
		client = o3.APIClientWithNEOPrivateNet()
	}

	response := client.GetNEOUTXO(address)
	if response.Code != 200 {
		return unspent, fmt.Errorf("Error cannot get utxo")
	}

	gasBalance := smartcontract.Balance{
		Amount: float64(0) / float64(100000000),
		UTXOs:  []smartcontract.UTXO{},
	}

	neoBalance := smartcontract.Balance{
		Amount: float64(0) / float64(100000000),
		UTXOs:  []smartcontract.UTXO{},
	}

	for _, v := range response.Result.Data {
		if strings.Contains(v.Asset, string(smartcontract.GAS)) {
			value, err := strconv.ParseFloat(v.Value, 64)
			if err != nil {
				continue
			}
			gasTX1 := smartcontract.UTXO{
				Index: v.Index,
				TXID:  v.Txid,
				Value: value,
			}
			gasBalance.UTXOs = append(gasBalance.UTXOs, gasTX1)
		}

		if strings.Contains(v.Asset, string(smartcontract.NEO)) {
			value, err := strconv.ParseFloat(v.Value, 64)
			if err != nil {
				continue
			}
			tx := smartcontract.UTXO{
				Index: v.Index,
				TXID:  v.Txid,
				Value: value,
			}
			neoBalance.UTXOs = append(neoBalance.UTXOs, tx)
		}
	}

	unspent.Assets[smartcontract.GAS] = &gasBalance
	unspent.Assets[smartcontract.NEO] = &neoBalance
	return unspent, nil
}

type RawTransaction struct {
	TXID string
	Data []byte
}

func MintTokensRawTransactionMobile(network string, scriptHash string, wif string, sendingAssetID string, amount float64, remark string, networkFeeAmountInGAS float64) (*RawTransaction, error) {
	rawTransaction := &RawTransaction{}
	fee := smartcontract.NetworkFeeAmount(networkFeeAmountInGAS)
	nep5 := UseNEP5WithNetworkFee(scriptHash, fee)
	wallet, err := GenerateFromWIF(wif)
	if err != nil {
		return nil, err
	}

	unspent, err := utxoFromO3Platform(network, wallet.Address)
	if err != nil {
		return nil, err
	}

	nativeAsset := smartcontract.NativeAssets[sendingAssetID]
	if nativeAsset == "" {
		return nil, fmt.Errorf("invalid assetID")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("Invalid amount. cannot be zero or less than zero")
	}

	data, txIDString, err := nep5.MintTokensRawTransaction(*wallet, nativeAsset, amount, unspent, remark)
	if err != nil {
		return nil, err
	}
	rawTransaction.Data = data
	rawTransaction.TXID = txIDString
	return rawTransaction, nil
}

func GenerateNEP6FromEncryptedKey(walletName, addressLabel, address, encryptedKey string) string {
	nep6Wallet := nep6.NewNEP6WithNEP2EncryptedKey(walletName, addressLabel, address, encryptedKey)
	b, err := json.Marshal(nep6Wallet)
	if err != nil {
		return ""
	}
	return string(b)
}

func SerializeTX(jsonString string) []byte {
	tx := NeonJSTransaction{}
	json.Unmarshal([]byte(jsonString), &tx)
	log.Printf("%+v", tx)
	final := NeonJSTXSerializer(tx)
	return final
}

func NEOAddresstoScriptHashBigEndian(neoAddress string) string {
	return NEOAddressToScriptHashWithEndian(neoAddress, binary.BigEndian)
}

func GetVarUInt(value int64) []byte {
	buff := new(bytes.Buffer)
	WriteVarUint(buff, uint64(value))

	return buff.Bytes()
}
