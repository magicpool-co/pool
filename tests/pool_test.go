//go:build integration

package tests

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/magicpool-co/pool/app/pool"
	"github.com/magicpool-co/pool/internal/log"
	"github.com/magicpool-co/pool/internal/node"
	"github.com/magicpool-co/pool/pkg/common"
	"github.com/magicpool-co/pool/pkg/stratum"
	"github.com/magicpool-co/pool/pkg/stratum/rpc"
)

type PoolSuite struct {
	suite.Suite
}

func (suite *PoolSuite) TestPool() {
	tests := []struct {
		chain     string
		priv      string
		opts      *pool.Options
		handshake []*rpc.Request
		requests  []*rpc.Request
		responses [][]byte
	}{
		{
			chain: "AE",
			priv: "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e" +
				"03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:          "AE",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     100000,
				ExtraNonceSize: 4,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("eth_submitHashrate",
					"0x000000000000000000000000025fc9f3",
					"0x8f4c405730375083a74b95eee0ff1ab1d472f176a13f115af1d553408a9add5d",
				),
				// a submission consists of:
				//	- worker id
				// 	- job id
				// 	- nonce
				//	- solution
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"356d5470",
					[]string{
						"01b89053", "01efc0e9", "01f007ae", "020a865c", "02e319c3", "0321d72a", "03e5f0a5", "04cbc625",
						"04fd1335", "0933185b", "09e1705b", "0c0008f7", "0c9ab2da", "0ecac105", "0f91a91a", "0fdba4a7",
						"1050b77d", "12a7cafd", "12adcca2", "1307ebfa", "145a0a29", "146a9b6b", "14b60e33", "152c4cf0",
						"156de5e0", "162ef9d1", "17487854", "19a5e13e", "1a8472ce", "1a8bee7d", "1aee1532", "1af23b43",
						"1c58d6db", "1c9a36d4", "1c9f5132", "1cbe6475", "1d89a910", "1e521b5c", "1e58619c", "1e5b84fe",
						"1e7afe96", "1f919ada",
					},
				),
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"85f9d18d",
					[]string{
						"01ffbd0b", "020072a2", "02b29b0c", "02ee4a27", "03f7a537", "0471bf13", "04806cd5", "0535b8ae",
						"062d5562", "064d8a83", "07a35c79", "081e5c37", "08a1f0b2", "08b21c7e", "09b25d57", "0a65d2dc",
						"0ae0572e", "0c9797e3", "0da570ea", "0e4a54e6", "1029f3c3", "103ce7a4", "118bb092", "11a84da4",
						"11d7df84", "13bf2ab3", "1487c882", "14f26ac7", "15357fa0", "157d6931", "15eaf0fe", "16541e8b",
						"1720ca3f", "17d30f85", "18226043", "18d4ed48", "191decd8", "19383ff0", "1974cae9", "1a90f244",
						"1cd82e94", "1f614b02",
					},
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"85f9d18d",
					[]string{
						"01ffbd0b", "020072a2", "02b29b0c", "02ee4a27", "03f7a537", "0471bf13", "04806cd5", "0535b8ae",
						"062d5562", "064d8a83", "07a35c79", "081e5c37", "08a1f0b2", "08b21c7e", "09b25d57", "0a65d2dc",
						"0ae0572e", "0c9797e3", "0da570ea", "0e4a54e6", "1029f3c3", "103ce7a4", "118bb092", "11a84da4",
						"11d7df84", "13bf2ab3", "1487c882", "14f26ac7", "15357fa0", "157d6931", "15eaf0fe", "16541e8b",
						"1720ca3f", "17d30f85", "18226043", "18d4ed48", "191decd8", "19383ff0", "1974cae9", "1a90f244",
						"1cd82e94", "1f614b02",
					},
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "CFX",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:         "CFX",
				PortDiffIdx:   map[int]int{0: 1},
				WindowSize:    100000,
				JobListSize:   10,
				PollingPeriod: time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe",
					"cfx:aajpuruxmg5z90x07z2ynt2u5wrknz717ymnu6mhdp.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				// a submission consists of:
				//	- worker id
				// 	- job id
				// 	- nonce
				//	- hash
				rpc.MustNewRequest("mining.submit",
					"cfx:aajpuruxmg5z90x07z2ynt2u5wrknz717ymnu6mhdp.worker",
					"000001",
					"0x7d444f1ed8ade6f0",
					"0xde5b0ae317379fa03f768eb102fb1e3671c9beaafecc52ae3c24eb77e80d6e03",
				),
				rpc.MustNewRequest("mining.submit",
					"cfx:aajpuruxmg5z90x07z2ynt2u5wrknz717ymnu6mhdp.worker",
					"000001",
					"0x7d444f1f4381257d",
					"0xde5b0ae317379fa03f768eb102fb1e3671c9beaafecc52ae3c24eb77e80d6e03",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"cfx:aajpuruxmg5z90x07z2ynt2u5wrknz717ymnu6mhdp.worker",
					"000001",
					"0x7d444f1f4381257d",
					"0xde5b0ae317379fa03f768eb102fb1e3671c9beaafecc52ae3c24eb77e80d6e03",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "CTXC",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:         "CTXC",
				PortDiffIdx:   map[int]int{0: 1},
				WindowSize:    100000,
				JobListSize:   10,
				PollingPeriod: time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("ctxc_submitLogin",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("ctxc_submitHashrate",
					"0x000000000000000000000000025fc9f3",
					"0x8f4c405730375083a74b95eee0ff1ab1d472f176a13f115af1d553408a9add5d",
				),
				// a submission consists of:
				// 	- nonce
				//	- header hash
				// 	- solution
				rpc.MustNewRequest("ctxc_submitWork",
					"0x100600000e6e4271",
					"0x3f62b7a620e9695597115ab1d5a219162136e7af6f187796b1a244793d27175d",
					"0x006695680524339d06262207067d42b707e9b7b1095dda3209a093390a4a4c2f0a88ba7c"+
						"0ab640eb0b7a3f820fdb43221018a35010a4be9613de7587151a4b78175247c719"+
						"5e191b1a9f9dc71ab585e7204acf8e2097713b212dc8c922a34ba5238d7e5423d7"+
						"6ae223e7589424062629240b63ab244d715927dbff102dfa4cf3332f58d83335f1"+
						"5f335149223599a0d23741a46d3936957e39839dc33b31a90f3c300d663f30870d",
				),
				rpc.MustNewRequest("ctxc_submitWork",
					"0x620600000e6e4271",
					"0x3f62b7a620e9695597115ab1d5a219162136e7af6f187796b1a244793d27175d",
					"0x0111504002d2191e02f87224048ab2e505ba0220087586cd08fb9667090137370b476013"+
						"0c34c0340c8a8ae90e90fe5a0e969ce90ec3e6681053873e105ab58d1288012f12"+
						"f1fd2714a3870d14f9c50c150afc73152a13ed15fdd43e162482961a46093d1c26"+
						"9c781ccb7cae1d42aacd1e3d381f2422df922448e69c2706079e27391d08277522"+
						"572a6679c13123b50132b9256e355ad16e387856523b8c56043b97bcc93dcd9484",
				),
				// test duplicate share
				rpc.MustNewRequest("ctxc_submitWork",
					"0x620600000e6e4271",
					"0x3f62b7a620e9695597115ab1d5a219162136e7af6f187796b1a244793d27175d",
					"0x0111504002d2191e02f87224048ab2e505ba0220087586cd08fb9667090137370b476013"+
						"0c34c0340c8a8ae90e90fe5a0e969ce90ec3e6681053873e105ab58d1288012f12"+
						"f1fd2714a3870d14f9c50c150afc73152a13ed15fdd43e162482961a46093d1c26"+
						"9c781ccb7cae1d42aacd1e3d381f2422df922448e69c2706079e27391d08277522"+
						"572a6679c13123b50132b9256e355ad16e387856523b8c56043b97bcc93dcd9484",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "ERG",
			priv:  "",
			opts: &pool.Options{
				Chain:                "ERG",
				PortDiffIdx:          map[int]int{0: 1},
				WindowSize:           100000,
				ExtraNonceSize:       2,
				JobListSize:          5,
				ForceErrorOnResponse: true,
				PollingPeriod:        time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- partial nonce (w/o extranonce)
				//	- unused
				// 	- full nonce
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"39132c4ae81f",
					"00000000",
					"ffff39132c4ae81f",
				),
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"39132f884e11",
					"00000000",
					"ffff39132f884e11",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
			},
		},
		{
			chain: "ETC",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:         "ETC",
				PortDiffIdx:   map[int]int{0: 1},
				WindowSize:    100000,
				JobListSize:   10,
				PollingPeriod: time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("eth_submitLogin",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("eth_submitHashrate",
					"0x000000000000000000000000025fc9f3",
					"0x8f4c405730375083a74b95eee0ff1ab1d472f176a13f115af1d553408a9add5d",
				),
				// a submission consists of:
				// 	- nonce
				//	- header hash
				// 	- mix digest
				rpc.MustNewRequest("eth_submitWork",
					"0x25a6eeb65f927295",
					"0x48b9e2560c8263614076d943d2c848044604b6e43c2423ed670ea5cc18b6edd8",
					"0x645fda1ed38a9884029f0533a68363447f7dc62916e19ea42a1259a44ce3017b",
				),
				rpc.MustNewRequest("eth_submitWork",
					"0x25a6eeb766e28d01",
					"0x48b9e2560c8263614076d943d2c848044604b6e43c2423ed670ea5cc18b6edd8",
					"0xefa4c0056216be6a82cfda1607d0aaa3ccfc484bad1eb7fbae3b7bd77d19a87a",
				),
				// test duplicate share
				rpc.MustNewRequest("eth_submitWork",
					"0x25a6eeb766e28d01",
					"0x48b9e2560c8263614076d943d2c848044604b6e43c2423ed670ea5cc18b6edd8",
					"0xefa4c0056216be6a82cfda1607d0aaa3ccfc484bad1eb7fbae3b7bd77d19a87a",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "FIRO",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:          "FIRO",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     300000,
				ExtraNonceSize: 1,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("eth_submitHashrate",
					"0x000000000000000000000000025fc9f3",
					"0x8f4c405730375083a74b95eee0ff1ab1d472f176a13f115af1d553408a9add5d",
				),
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- nonce
				//	- header hash
				// 	- mix digest
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"0xff00000049466d8c",
					"0x93f52026533c86a3797637f6b82c96b99c90ce68b4649cac0d5af649df20c410",
					"0xbdca50daa1912a3b826196f3115b3ef5e6060efd6b38ccc09992bfecdcb85403",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"0xff00000049466d8c",
					"0x93f52026533c86a3797637f6b82c96b99c90ce68b4649cac0d5af649df20c410",
					"0xbdca50daa1912a3b826196f3115b3ef5e6060efd6b38ccc09992bfecdcb85403",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "FLUX",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:          "FLUX",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     100000,
				ExtraNonceSize: 4,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("eth_submitHashrate",
					"0x000000000000000000000000025fc9f3",
					"0x8f4c405730375083a74b95eee0ff1ab1d472f176a13f115af1d553408a9add5d",
				),
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- time (unused)
				//	- nonce
				// 	- solution
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"bde70e63",
					"000000000000000000000000000000000000000000000000cd000000",
					"3412083e673b31323c60778f6bad22dde2e8c45ce9179603a4e625273da579a61de2c052c5ef509527b2b99d36f1f460e509eae779",
				),
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"bde70e63",
					"0000000000000000000000000000000000000000000000002e200000",
					"3401a7ef31e4b2a3b12ef557f15628631e3cbb2e2cdd9d778583f01354c86f991dfbb6107bec3b4732369cccbe1584798a2b2bc81a",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"bde70e63",
					"0000000000000000000000000000000000000000000000002e200000",
					"3401a7ef31e4b2a3b12ef557f15628631e3cbb2e2cdd9d778583f01354c86f991dfbb6107bec3b4732369cccbe1584798a2b2bc81a",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "KAS",
			priv:  "9476ca4050e719e3fb958be7ee64016d751e22d0063cca6b13880284c5bb42ad",
			opts: &pool.Options{
				Chain:           "KAS",
				PortDiffIdx:     map[int]int{0: 1},
				WindowSize:      100000,
				ExtraNonceSize:  2,
				JobListSize:     100,
				JobListAgeLimit: 12,
				PollingPeriod:   time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"kaspa:qyp4ek94qc9k7aqzmpe4l7kdp6pvqus3gqehy89zdlc9dssvhc2rqjq2wr26hvd.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- nonce
				rpc.MustNewRequest("mining.submit",
					"kaspa:qyp4ek94qc9k7aqzmpe4l7kdp6pvqus3gqehy89zdlc9dssvhc2rqjq2wr26hvd.worker",
					"000001",
					"ffff6a003aa3487c",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"kaspa:qyp4ek94qc9k7aqzmpe4l7kdp6pvqus3gqehy89zdlc9dssvhc2rqjq2wr26hvd.worker",
					"000001",
					"ffff6a003aa3487c",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "NEXA",
			priv:  "9476ca4050e719e3fb958be7ee64016d751e22d0063cca6b13880284c5bb42ad",
			opts: &pool.Options{
				Chain:          "NEXA",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     100000,
				ExtraNonceSize: 8,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"nexa:qzrpqsursz4dprxly2zfrvh8qgu64zvfe55dqxj2gy.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- nonce
				//  - timestamp
				rpc.MustNewRequest("mining.submit",
					"nexa:qzrpqsursz4dprxly2zfrvh8qgu64zvfe55dqxj2gy.worker",
					"1",
					"ffffffffa60c1c760017f000",
					"000000006425fe9e",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"nexa:qzrpqsursz4dprxly2zfrvh8qgu64zvfe55dqxj2gy.worker",
					"1",
					"ffffffffa60c1c760017f000",
					"0000000064155638",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
		{
			chain: "NEXA", // wildrig testing
			priv:  "9476ca4050e719e3fb958be7ee64016d751e22d0063cca6b13880284c5bb42ad",
			opts: &pool.Options{
				Chain:          "NEXA",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     100000,
				ExtraNonceSize: 8,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe", "WildRig/0.36.6L beta"),
				rpc.MustNewRequest("mining.authorize",
					"nexa:qzrpqsursz4dprxly2zfrvh8qgu64zvfe55dqxj2gy.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- extraNonce1
				//  - timestamp (?, unused anyways)
				//	- extraNonce2
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"nexa:qzrpqsursz4dprxly2zfrvh8qgu64zvfe55dqxj2gy.worker",
					"1",
					"ffffffff",
					"0c6f67a3",
					"3be1a62200000000",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
			},
		},
		{
			chain: "RVN",
			priv:  "03620b2ed304234abe4f02e4f95ece19626989351487c0f93821e4827ed1301e",
			opts: &pool.Options{
				Chain:          "RVN",
				PortDiffIdx:    map[int]int{0: 1},
				WindowSize:     300000,
				ExtraNonceSize: 1,
				JobListSize:    5,
				PollingPeriod:  time.Millisecond * 100,
			},
			handshake: []*rpc.Request{
				rpc.MustNewRequest("mining.subscribe"),
				rpc.MustNewRequest("mining.authorize",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"x",
				),
			},
			requests: []*rpc.Request{
				rpc.MustNewRequest("eth_submitHashrate",
					"0x00000000000000000000000000cf41fb",
					"0x456f26403e3042c1b482aad767dacfc70415bfff3020786a3db35bf8e44b3e0a",
				),
				// a submission consists of:
				// 	- worker id
				//	- job id
				// 	- nonce
				//	- header hash
				// 	- mix digest
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"0xff5af6135c7d5d01",
					"0x6fc2495aa1c4e6a90d7f5639c67dc3334647b8c41ef42a1a1cd690e49fe9e7f1",
					"0xf0587f05a6dfbac45f1d2d39fd2f3eb43639555e42224e9173757618baa2329f",
				),
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"0xff5af611df410db8",
					"0x6fc2495aa1c4e6a90d7f5639c67dc3334647b8c41ef42a1a1cd690e49fe9e7f1",
					"0xee1fae60fca1ea2b42195cf279ecff6ec62d1f60f7048296d9b83548e3ec05ba",
				),
				// test duplicate share
				rpc.MustNewRequest("mining.submit",
					"ETH:0x0000000000000000000000000000000000000000.worker",
					"000001",
					"0xff5af611df410db8",
					"0x6fc2495aa1c4e6a90d7f5639c67dc3334647b8c41ef42a1a1cd690e49fe9e7f1",
					"0xee1fae60fca1ea2b42195cf279ecff6ec62d1f60f7048296d9b83548e3ec05ba",
				),
			},
			responses: [][]byte{
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(true),
				common.MustMarshalJSON(false),
			},
		},
	}

	logger, err := log.New(map[string]string{"LOG_LEVEL": "ERROR"}, "pooltest", nil)
	if err != nil {
		suite.T().Errorf("failed to create logger: %v", err)
		return
	}

	for i, tt := range tests {
		miningNode, err := node.GetMiningNode(true, tt.chain, tt.priv, nil, logger, nil)
		if err != nil {
			suite.T().Errorf("failed to create node: %d: %s: %v", i, tt.chain, err)
		}

		server, err := pool.New(miningNode, pooldbClient, redisClient, logger, nil, nil, tt.opts)
		if err != nil {
			suite.T().Errorf("failed to create server: %d: %s: %v", i, tt.chain, err)
			continue
		}

		func() {
			go server.Serve()
			defer server.Stop()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// have to wait for server to start to properly instantiate port
			// since it's unknown (it's as zero and go chooses it)
			time.Sleep(time.Millisecond * 100)
			client := stratum.NewClient(ctx, fmt.Sprintf("localhost:%d", server.Port(0)), time.Second*5, time.Second)
			reqCh, resCh, errCh := client.Start(tt.handshake)
			time.Sleep(time.Millisecond * 100)

			var ready, started, completed bool
			for {
				if completed {
					return
				}

				select {
				case <-time.After(time.Second * 6):
					if !completed {
						suite.T().Errorf("failed on %d: %s: never ran test requests", i, tt.chain)
					}
					return
				case <-reqCh:
					ready = true
				case <-resCh:
					ready = true
				case err := <-errCh:
					suite.T().Errorf("received error: %d: %s: %v", i, tt.chain, err)
				}

				if !ready || started {
					continue
				}
				started = true

				go func() {
					for j, req := range tt.requests {
						res, err := client.WriteRequest(req)
						if err != nil {
							suite.T().Errorf("failed to write request: %d: %s: %s: %v", i, tt.chain, req.Method, err)
							continue
						} else if bytes.Compare(res.Result, tt.responses[j]) != 0 {
							suite.T().Errorf("failed on response: mismatch: %d: %s: have %s, want %s",
								i, tt.chain, res.Result, tt.responses[j])
						}
					}
					completed = true
				}()
			}
		}()
	}
}
