package account_scraper

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client"
	"github.com/centrifuge/go-substrate-rpc-client/scale"
	"github.com/centrifuge/go-substrate-rpc-client/types"
)

func processRange(api *gsrpc.SubstrateAPI, meta *types.Metadata, key types.StorageKey, lower, upper uint64, accountSet map[types.AccountID]bool) error {
	fmt.Printf("Processing %d - %d\n", lower, upper)

	lbh, err := api.RPC.Chain.GetBlockHash(lower)
	if err != nil {
		return err
	}

	ubh, err := api.RPC.Chain.GetBlockHash(upper)
	if err != nil {
		return err
	}

	rawSet, err := api.RPC.State.QueryStorage([]types.StorageKey{key}, lbh, ubh)
	if err != nil {
		return err
	}

	for i := 0 ; i < len(rawSet) ; i++ {
		for j := 0; j < len(rawSet[i].Changes); j++ {
			raw := rawSet[i].Changes[j].StorageData
			events := EventRecords{}
			err = types.EventRecordsRaw(raw).DecodeEventRecords(meta, &events)
			if err != nil {
				bb, err1 := api.RPC.Chain.GetBlock(rawSet[i].Block)
				if err1 != nil {
					fmt.Printf("Unexpected error getting block hash %s: %s\n", rawSet[i].Block.Hex(), err.Error())
					break
				}
				fmt.Printf("Error processing events in block %d with error %s\n", bb.Block.Header.Number, err.Error())
				break
			}
			if len(events.Balances_Endowed) > 0 {
				for k := 0; k < len(events.Balances_Endowed); k++ {
					fmt.Printf("%x\n", events.Balances_Endowed[k].Who)
					accountSet[events.Balances_Endowed[k].Who] = true
				}
			}
		}
	}

	return nil
}

func encodeAndSave(accountSet map[types.AccountID]bool) error {
	var accounts []types.AccountID
	for key, _ := range accountSet {
		accounts = append(accounts, key)
	}

	var buffer = bytes.Buffer{}
	err := scale.NewEncoder(&buffer).Encode(accounts)
	if err != nil {
		return err
	}

	_, err = os.Stat("build")

	if os.IsNotExist(err) {
		errDir := os.MkdirAll("build", 0755)
		if errDir != nil {
			log.Fatal(err)
		}

	}

	f, err := os.Create("build/accounts.scale")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func sanityCheck() error {
	dataRead, err := ioutil.ReadFile("build/accounts.scale")
	if err != nil {
		return err
	}

	var readAccounts []types.AccountID
	err = types.DecodeFromBytes(dataRead, &readAccounts)
	if err != nil {
		return err
	}


	fmt.Println("Accounts Identified:")
	for x := 0; x < len(readAccounts); x++ {
		fmt.Printf("%x\n", readAccounts[x])
	}

	return nil
}

func Process(targetURL string) error {
	//targetURL = "wss://fullnode-archive.centrifuge.io"
	api, err := gsrpc.NewSubstrateAPI(targetURL)
	if err != nil {
		return err
	}

	meta, err := api.RPC.State.GetMetadataLatest()
	if err != nil {
		return err
	}

	key, err := types.CreateStorageKey(meta, "System", "Events", nil, nil)
	if err != nil {
		return err
	}

	step := uint64(5000)
	latest, err := api.RPC.Chain.GetBlockLatest()
	if err != nil {
		return err
	}

	latestNumber := uint64(latest.Block.Header.Number)
	//latestNumber := uint64(2304153)

	fmt.Println("Processing blocks until", latestNumber)

	accountSet := genesisAccounts()
	for i := uint64(0); i < latestNumber; i+=step {
		lower := i
		upper := i + step
		if upper > latestNumber {
			upper = latestNumber
		}

		err := processRange(api, meta, key, lower, upper, accountSet)
		if err != nil {
			return errors.Wrap(err, "Error Processing Range")
		}
	}

	err = encodeAndSave(accountSet)
	if err != nil {
		return errors.Wrap(err, "Error Encoding/Saving")
	}

	err = sanityCheck()
	if err != nil {
		return errors.Wrap(err, "Error Sanity Check")
	}

	return nil
}

func genesisAccounts() map[types.AccountID]bool {
	accs := []string{
		"0x123bdd258d11c2afb5cc2aaf116abc113db2b99b90bff6864cbc12fb0d9e7a7c",
		"0xba2c4540acac96a93e611ec4258ce05338434f12107d35f29783bbd2477dd20e",
		"0xaa35391992c3ae5effb7b347db468ce3015e0fd61db940af96adab4f420d775f",
		"0xd893ea3ee82b684574a124eae65ba6e5f8edba4448ff90ed19a83d42e00d9a03",
		"0xe86577546f1981927ca81364b7ccf8566a2056fd162e9b6a146dc52afdd88c50",
		"0x3e70df435b1c0535b13939870e0a668f5f51482ac7912fcb7ee7f0fd93a69a38",
		"0xa4d9c40a082074ad257d5913c7c0adc671f7e8549aabb30f8c5eae7adfbd9916",
		"0x7e3a27ebc30843a9b856bcd77423bd10db0dd98caa295e4dbe87783dcfd3e939",
		"0xa4dd9a830a1fb478f6e1a569a782d10d0caf25505fb442ac560df7370db7382e",
		"0xba537b319fd88d968cc19230b6ba734fc46989b8b93e267a6f5577b2038b2374",
		"0x563d11af91b3a166d07110bb49e84094f38364ef39c43a26066ca123a8b9532b",
		"0x242f0781faa44f34ddcbc9e731d0ddb51c97f5b58bb2202090a3a1c679fc4c63",
		"0x1069677e1538e6b56c8d96570f43383194e5ea178d34531dd8d7e44adcc7a773",
		"0xa8ec81084413b23c6b933fb62300361093a18cb37457c88d3db9813714fd9541",
		"0x6a81eca57b0e8cbfbbfde6d17d5c72ebcf884c4935e49695858ca6d0fa899516",
		"0x72bf1d2703848ebc54a208241849b80fda765b786e9d4ef4fbf04e1f05c43e3c",
		"0x88f63c7a1adb15e0fdee166684f190afb19943b065fbbe037271593b8403454f",
		"0x5eb623df668fd9d255711f7a055c303293b29d120be7455c24d18dba0d32691f",
		"0x663f032d66e8e7c70270e57dfe9500a64efa20f5239fd5e632de86ec8d7fc30f",
		"0xacdbc2ab1dd9274a5d0699a9b666d531b880aef033fd748e5e09522ac5896010",
		"0xa09f04e82e2ff56d3f9e6de4b616d7025f712b14ed1cc33a40243f15bf6e3644",
		"0x70958392b5dfbca9690a0dae3a40cccbe42458fb9d048410ca581fd206128d05",
		"0xe4f6fb83c8cc774c64a26e513b2ae4db197478dd68bc460e7fe386ebd41f091a",
		"0x7c0072c14b3788e2f4a942f4f989945bc3417f0af1215f5b0f9590882daae908",
		"0x6e03bb0c55d624ac6365f51584918da27697b5dde64f34ab700e97fc3ce14b50",
		"0x8a4e6b8b355c950aa84e3c988fc0059c10651513e5ce2181285912faa7fedf1a",
		"0x6d6f646c63622f62726964670000000000000000000000000000000000000000",
		"0x4bf08e42ca674b08ff58c8ae3b4c08142d07be86cfb689f0c2ab98ab8a6f475c",
		"0xed607b21c31d1bec51f19361bfa63d21d6e3ae8539a06c8c68e37ccd9b481f96",
		"0xb13e8db4ff76cc2ee92b94bc160a087db9ad60e18dbb0d466d05f4c7a8891b3e",
		"0x824b25de84e2200f3366df34e962e4babc4ddbc1d03ca8c1d6532e2812c709ec",
		"0x5e0030c3a79685ee1ee4130ffcc9ff641d35685c20e2b3255debec62a3de8a2e",
		"0x22969478cdf372b63a48e216ce0760af42cc119a4024835ac81cea58c265bb31",
		"0xf4a6f480013b9f4d2314d2b8bd05b54462fc034a46430591abe9760dcd000d6c",
		"0x9c664d8d01723bab456e618b2f9e64f1e0bd47adc858f94eeb3a2229118d5c15",
		"0x5cc4dc422fe3132544f11ee65ad0f91e46b4eb633944e0070d5f098189aae976",
		"0x304f01999aa7209e9ff0bb0467b5e8150776800d887104d980a658a50fed4325",
		"0xdc638c981f9cc72b1d36233dc74c20283c12ec3b0bd560255e564976b6fa3611",
		"0x987462947334b60f037bd132e17f2887330c45a0445bff5ba66a78cfcbcb1b49",
		"0x462186719384f6d9ae2afca0664acfc4ff4e06ab4a0fc975abcbf9ffaedcc975",
		"0xd005f367444aabf3bd902fae60367dedac9dc2409f82df518132d2f22023a83f",
		"0x3df9212d35f9ac374340f6baa51653a995578b9fb4f2d7be22277cd582382d18",
		"0x36214a18301fb1c724cda7ded6d32743a1871e13d0e7b1c76da24ee0a97a8238",
		"0xb469b015ca12cbdeebac0fa76f252870e7d5ff2dd3772b6a8b716b5f707ab016",
		"0xee404a8f1c53e6f0ea11657028df38bf829c021cc1fe017f979350c9f1f98579",
		"0x2ce2ce3bd7ff2759b048e96a637fea72d5856519a341b8fa6fa0b766b67adf32",
		"0x5879eaaf916be4413f10536a4b9a3c6609bed89604dd28becfdbccfe2fee8357",
		"0xd8e240b073ada6fa8ee7e4d4c6876a4a232172c5d232351cb31feb3f082a4441",
		"0x580fbb8924fc129787e700a527edb6c4ae04f605335358e8e560faac00453835",
		"0x64771c6903e2705733396307ab641b67057220c9fc4d3a49212f4d1e18ed6d03",
		"0x66f14fb7ee47a59498c3f692b8b54a522efad7aa1fbe7836e99eb1b721d0805c",
		"0x3ab513f457cf3f77c5235a1b34f29cb96ab2454d6d0d2b94e22a743a8e7f8731",
	}

	var accountSet = make(map[types.AccountID]bool)
	for _, elem := range accs {
		accountSet[types.NewAccountID(hexutil.MustDecode(elem))] = true
	}

	return accountSet
}

