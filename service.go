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

func loadAccounts() (map[types.AccountID]bool, error) {
	dataRead, err := ioutil.ReadFile("build/accounts.scale")
	if err != nil {
		return nil, err
	}

	var listAccounts []types.AccountID
	err = types.DecodeFromBytes(dataRead, &listAccounts)
	if err != nil {
		return nil, err
	}

	var mapAccounts = make(map[types.AccountID]bool)
	for _, elem := range listAccounts {
		mapAccounts[elem] = true
	}

	return mapAccounts, nil
}

func Process(targetURL string, append bool) error {
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

	var accountSet = make(map[types.AccountID]bool)
	if append {
		fmt.Println("Appending to existing Accounts File")
		accountSet, err = loadAccounts()
		if err != nil {
			return err
		}
	}

	addTestAccounts(accountSet)
	addGenesisAccounts(accountSet)

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

	// Sanity Check
	readAccounts, err := loadAccounts()
	if err != nil {
		return errors.Wrap(err, "Error Sanity Check")
	}

	fmt.Println("Accounts Identified:")
	for key, _ := range readAccounts {
		fmt.Printf("%x\n", key)
	}

	return nil
}

func addGenesisAccounts(accountSet map[types.AccountID]bool) {
	mainAccs := []string{
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

	amberAccs := []string{
		"0xb61f93a69fc0f7fb0f0e390def11a05b365f52cc2d76c8aaf6c6d1ccf8868d51",
		"0xc4461198656800fbaf42331ddf9394dd4d6233f843481c82cd44ff134640253b",
		"0x1af412c1fd789de98532e890828b42b71941a23dd3ae659a4657f0c287a2c620",
		"0x9a6474cf589a2fff75ac6fabcf0756bb86581e0e777e4da7b1c34d1c25003b6d",
		"0xc697db2284c9d2938b59ab34d4a39fc98b7e75a4a53aaf64df9f923b2da79943",
		"0x90d93f5cfdad8eb8bdf699a49f52aaf6ff45e2097f035201f7e2fb62f8ff1a59",
		"0x60e25d10ef42645a5b9e41b82b5354053c15c5a9066b2f0857819505c73a1c18",
		"0x58473270bfe850d36c8a9b17851cdaab6eaef34e8a19f203032a841b17e6225a",
		"0x281a3a7c24a57383c8ab210cf68c77809c59c40bf561e1d273551fd04b0bf003",
		"0x6c5d6e0a1616138d429ec82343c64e19ca91fdc4fc83c6230291c6629ee31e5c",
		"0x48415841c8876bb0ae0d0c5c19f4113d20759ce5417f5755ff858aef1c8d5847",
		"0x806583290bf6a8f96a69e092939555cedcdde7aacd8224bee32cf16316fa005f",
		"0x58eff1cf80796776dca1ffc26983c905ec35bc298f5f2e694fce682564c07f51",
		"0xbec5c3a4d94bf1fdf93daad479c4733ef4775232e4bbe48d7907fbe7f2b77d46",
		"0x42bbdad494b897fafd9bc235c8f8aa81a95f6681188eba6d4ac89669f118563a",
		"0x308f4b699d3ba6583b26a34165fe3759d082a91d09e66254796101fdfc17a370",
		"0x5afd1ff6fd10fff5124c4c36bc96355c89d9bef7f11ac8fbe336e9eed5056237",
		"0x6c831f845da5d2cba3d47844ae1a39f9d466a16e39cd4e24a29797113f4a1349",
		"0xba4b3e752a0737ce22af2b44a8a2e893ccb491644715de57ade77b6d270fea5f",
		"0x22698426dae9285b03f77eb2f8bb079980cd7efe107adf0d894a250219a5e40d",
		"0x4cc2abff7349d0ec56e8c00cc4d3250dc0319b89d2940292c24dad35528b5031",
		"0xcc720f4fca68808d1a6c7ddbbac58dfdcfca9a12b2ad221235aebebb5d91b468",
		"0x86e902161931aea76200686f51fc2a4135a53e419eb70d864e5e45858569c458",
		"0x2090ba294eda76dbf1fe53a1292373198bbe140d165c7a3718a3cf483ced6203",
		"0x5ae1ab6d1fffe69e07bae35aa873beb9f1a4352134629535ddcb0a9bc5313974",
		"0xfe74f018297259cbdebbf58a5e755f4d74b95b4d45244f736493a7a195e6c14c",
		"0xce291beba4e958c935b3dff2c04df16c773ae1949d5204328365b7e4aa2f5049",
		"0x00b2a45da53f66199472a3e3e096fc1174aca8ad06c42648268b4c16bea61b69",
		"0xa665dd831865ab9e210fff87379566c589e50dcd88a5b90cb327aedc3307ec02",
		"0x6c0b99e13f2a644186c64cae0a04497d61b8d6cca2146f7f175d3953ee72c769",
		"0x4cc2d92249624cd61bcbcfeb8fbb5a8de7699feedb1b3cf4c161f9e14faeeb37",
		"0x4bf08e42ca674b08ff58c8ae3b4c08142d07be86cfb689f0c2ab98ab8a6f475c",
	}

	flintAccs := []string{
		"0xc4051f94a879bd014647993acb2d52c4059a872b6e202e70c3121212416c5842",
		"0xe85164fc14c1275c398301fbfb9663916f4b0847331aa8ab2097c79358cb2a3d",
		"0x6c8f1e49c090d4998b23cc68d52453563785df4e84f3a10024b77d8b4649d51c",
		"0xa665dd831865ab9e210fff87379566c589e50dcd88a5b90cb327aedc3307ec02",
		"0x6c0b99e13f2a644186c64cae0a04497d61b8d6cca2146f7f175d3953ee72c769",
		"0x4cc2d92249624cd61bcbcfeb8fbb5a8de7699feedb1b3cf4c161f9e14faeeb37",
		"0x4bf08e42ca674b08ff58c8ae3b4c08142d07be86cfb689f0c2ab98ab8a6f475c",
	}

	for _, elem := range mainAccs {
		accountSet[types.NewAccountID(hexutil.MustDecode(elem))] = true
	}
	for _, elem := range amberAccs {
		accountSet[types.NewAccountID(hexutil.MustDecode(elem))] = true
	}
	for _, elem := range flintAccs {
		accountSet[types.NewAccountID(hexutil.MustDecode(elem))] = true
	}

	return
}

func addTestAccounts(accountSet map[types.AccountID]bool) {
	accs := []string{
		"0xc2cda5af8590d296eff5d7bb3ddf8235ca0f220743861808d611f4d5e5c120f8",
		"0x20caaa19510a791d1f3799dac19f170938aeb0e58c3d1ebf07010532e599d728",
		"0x9efc9f132428d21268710181fe4315e1a02d838e0e5239fe45599f54310a7c34",
		"0xc405224448dcd4259816b09cfedbd8df0e6796b16286ea18efa2d6343da5992e",
		"0xa23153e26c377a172c803e35711257c638e6944ad0c0627db9e3fc63d8503639",
		"0x8f9f7766fb5f36aeeed7a05b5676c14ae7c13043e3079b8a850131784b6d15d8",
		"0x42a6fcd852ef2fe2205de2a3d555e076353b711800c6b59aef67c7c7c1acf04d",
		"0xbe1ce959980b786c35e521eebece9d4fe55c41385637d117aa492211eeca7c3d",
	}
	for _, elem := range accs {
		accountSet[types.NewAccountID(hexutil.MustDecode(elem))] = true
	}
}

