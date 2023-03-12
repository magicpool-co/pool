package accounting

import (
	"math/big"
	"reflect"
	"testing"
)

func TestSplitValue(t *testing.T) {
	tests := []struct {
		value     *big.Int
		idx       map[uint64]uint64
		values    map[uint64]*big.Int
		remainder *big.Int
	}{
		{
			value: new(big.Int).SetUint64(1992800000000000000),
			idx: map[uint64]uint64{
				1: 5, 2: 1367,
			},
			values: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(7262390670553935), 2: new(big.Int).SetUint64(1985537609329446064),
			},
			remainder: new(big.Int).SetUint64(1),
		},
	}

	for i, tt := range tests {
		values, remainder, err := splitValue(tt.value, tt.idx)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !reflect.DeepEqual(values, tt.values) {
			t.Errorf("failed on %d: values mismatch: have %v, want %v", i, values, tt.values)
		} else if remainder.Cmp(tt.remainder) != 0 {
			t.Errorf("failed on %d: remainder mismatch: have %s, want %s", i, remainder, tt.remainder)
		}
	}
}

func TestCreditRound(t *testing.T) {
	tests := []struct {
		roundValue   *big.Int
		minerIdx     map[uint64]uint64
		recipientIdx map[uint64]uint64
		outputValues map[uint64]*big.Int
		outputFees   map[uint64]*big.Int
	}{
		{
			roundValue: new(big.Int).SetUint64(1992800000000000000),
			minerIdx: map[uint64]uint64{
				1: 5, 2: 1367,
			},
			recipientIdx: map[uint64]uint64{
				500: 50,
				501: 50,
			},
			outputValues: map[uint64]*big.Int{
				// miners
				1: new(big.Int).SetUint64(7261664431486880), 2: new(big.Int).SetUint64(1985339055568513119),
				// recipients
				500: new(big.Int).SetUint64(99640000000001), 501: new(big.Int).SetUint64(99640000000000),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(726239067055), 2: new(big.Int).SetUint64(198553760932945),
			},
		},
		{
			roundValue: new(big.Int).SetUint64(250002379571),
			minerIdx: map[uint64]uint64{
				1: 106019, 2: 34068, 3: 20906, 4: 16317,
				5: 15928, 6: 13701, 7: 13054, 8: 13045,
				9: 11234, 10: 11133, 11: 7250, 12: 6155,
				13: 5746, 14: 4540, 15: 4523, 16: 4111,
				17: 3944, 18: 3899, 19: 2714, 20: 1713,
			},
			recipientIdx: map[uint64]uint64{
				500: 100,
			},
			outputValues: map[uint64]*big.Int{
				// miners
				1: new(big.Int).SetUint64(88341172598), 2: new(big.Int).SetUint64(28387431197),
				3: new(big.Int).SetUint64(17420090307), 4: new(big.Int).SetUint64(13596269661),
				5: new(big.Int).SetUint64(13272132326), 6: new(big.Int).SetUint64(11416466914),
				7: new(big.Int).SetUint64(10877349032), 8: new(big.Int).SetUint64(10869849711),
				9: new(big.Int).SetUint64(9360819598), 10: new(big.Int).SetUint64(9276660547),
				11: new(big.Int).SetUint64(6041120000), 12: new(big.Int).SetUint64(5128702566),
				13: new(big.Int).SetUint64(4787900072), 14: new(big.Int).SetUint64(3782991007),
				15: new(big.Int).SetUint64(3768825622), 16: new(big.Int).SetUint64(3425523354),
				17: new(big.Int).SetUint64(3286369280), 18: new(big.Int).SetUint64(3248872673),
				19: new(big.Int).SetUint64(2261462025), 20: new(big.Int).SetUint64(1427370835),
				// recipients
				500: new(big.Int).SetUint64(25000246),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(8835001), 2: new(big.Int).SetUint64(2839027),
				3: new(big.Int).SetUint64(1742184), 4: new(big.Int).SetUint64(1359763),
				5: new(big.Int).SetUint64(1327346), 6: new(big.Int).SetUint64(1141761),
				7: new(big.Int).SetUint64(1087844), 8: new(big.Int).SetUint64(1087094),
				9: new(big.Int).SetUint64(936175), 10: new(big.Int).SetUint64(927758),
				11: new(big.Int).SetUint64(604172), 12: new(big.Int).SetUint64(512921),
				13: new(big.Int).SetUint64(478838), 14: new(big.Int).SetUint64(378337),
				15: new(big.Int).SetUint64(376920), 16: new(big.Int).SetUint64(342587),
				17: new(big.Int).SetUint64(328670), 18: new(big.Int).SetUint64(324920),
				19: new(big.Int).SetUint64(226168), 20: new(big.Int).SetUint64(142752),
			},
		},
		{
			roundValue: new(big.Int).SetUint64(2042523002164311183),
			minerIdx: map[uint64]uint64{
				1: 854411, 2: 695607, 3: 78227, 4: 22131,
				5: 18989, 6: 16641, 7: 15837, 8: 14915,
				9: 11140, 10: 10781, 11: 6347, 12: 6222,
				13: 5450, 14: 5254, 15: 4962, 16: 4809,
				17: 3438, 18: 3251, 19: 2986, 20: 2536,
				21: 2398, 22: 2356, 23: 2350, 24: 1342,
				25: 1289, 26: 1260, 27: 1162, 28: 1138,
				29: 1099, 30: 1084, 31: 1080, 32: 1055,
				33: 1050, 34: 1032, 35: 1006, 36: 189,
				37: 182, 38: 161, 39: 140, 40: 137,
				41: 136, 42: 126, 43: 123, 44: 122,
				45: 116, 46: 70, 47: 66, 48: 61,
				49: 59,
			},
			recipientIdx: map[uint64]uint64{
				500: 50,
				501: 50,
			},
			outputValues: map[uint64]*big.Int{
				// miners

				1: new(big.Int).SetUint64(966039631555447758), 2: new(big.Int).SetUint64(786487919733465918),
				3: new(big.Int).SetUint64(88447342388719260), 4: new(big.Int).SetUint64(25022410860760938),
				5: new(big.Int).SetUint64(21469909169716210), 6: new(big.Int).SetUint64(18815143424785268),
				7: new(big.Int).SetUint64(17906100980609596), 8: new(big.Int).SetUint64(16863641859303664),
				9: new(big.Int).SetUint64(12595438840941523), 10: new(big.Int).SetUint64(12189535560519799),
				11: new(big.Int).SetUint64(7176234319879340), 12: new(big.Int).SetUint64(7034903094105759),
				13: new(big.Int).SetUint64(6162041443728124), 14: new(big.Int).SetUint64(5940434081715149),
				15: new(big.Int).SetUint64(5610284338308064), 16: new(big.Int).SetUint64(5437294917961201),
				17: new(big.Int).SetUint64(3887174033676567), 18: new(big.Int).SetUint64(3675742519919290),
				19: new(big.Int).SetUint64(3376120321279298), 20: new(big.Int).SetUint64(2867327908494407),
				21: new(big.Int).SetUint64(2711298235240374), 22: new(big.Int).SetUint64(2663810943380451),
				23: new(big.Int).SetUint64(2657027044543319), 24: new(big.Int).SetUint64(1517332039905163),
				25: new(big.Int).SetUint64(1457407600177165), 26: new(big.Int).SetUint64(1424618755797694),
				27: new(big.Int).SetUint64(1313815074791207), 28: new(big.Int).SetUint64(1286679479442679),
				29: new(big.Int).SetUint64(1242584137001322), 30: new(big.Int).SetUint64(1225624389908492),
				31: new(big.Int).SetUint64(1221101790683738), 32: new(big.Int).SetUint64(1192835545529022),
				33: new(big.Int).SetUint64(1187182296498078), 34: new(big.Int).SetUint64(1166830599986683),
				35: new(big.Int).SetUint64(1137433705025778), 36: new(big.Int).SetUint64(213692813369654),
				37: new(big.Int).SetUint64(205778264726333), 38: new(big.Int).SetUint64(182034618796372),
				39: new(big.Int).SetUint64(158290972866410), 40: new(big.Int).SetUint64(154899023447844),
				41: new(big.Int).SetUint64(153768373641655), 42: new(big.Int).SetUint64(142461875579769),
				43: new(big.Int).SetUint64(139069926161203), 44: new(big.Int).SetUint64(137939276355014),
				45: new(big.Int).SetUint64(131155377517883), 46: new(big.Int).SetUint64(79145486433205),
				47: new(big.Int).SetUint64(74622887208450), 48: new(big.Int).SetUint64(68969638177507),
				49: new(big.Int).SetUint64(66708338565130),
				// recipients
				500: new(big.Int).SetUint64(102126150108228), 501: new(big.Int).SetUint64(102126150108228),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(96613624517996), 2: new(big.Int).SetUint64(78656657639110),
				3: new(big.Int).SetUint64(8845618800752), 4: new(big.Int).SetUint64(2502491335210),
				5: new(big.Int).SetUint64(2147205637536), 6: new(big.Int).SetUint64(1881702512730),
				7: new(big.Int).SetUint64(1790789176979), 8: new(big.Int).SetUint64(1686532839214),
				9: new(big.Int).SetUint64(1259669851079), 10: new(big.Int).SetUint64(1219075463598),
				11: new(big.Int).SetUint64(717695201508), 12: new(big.Int).SetUint64(703560665477),
				13: new(big.Int).SetUint64(616265770950), 14: new(big.Int).SetUint64(594102818453),
				15: new(big.Int).SetUint64(561084542285), 16: new(big.Int).SetUint64(543783870183),
				17: new(big.Int).SetUint64(388756278995), 18: new(big.Int).SetUint64(367611013093),
				19: new(big.Int).SetUint64(337645796708), 20: new(big.Int).SetUint64(286761466997),
				21: new(big.Int).SetUint64(271156939218), 22: new(big.Int).SetUint64(266407735112),
				23: new(big.Int).SetUint64(265729277382), 24: new(big.Int).SetUint64(151748378829),
				25: new(big.Int).SetUint64(145755335551), 26: new(big.Int).SetUint64(142476123192),
				27: new(big.Int).SetUint64(131394646944), 28: new(big.Int).SetUint64(128680816026),
				29: new(big.Int).SetUint64(124270840784), 30: new(big.Int).SetUint64(122574696461),
				31: new(big.Int).SetUint64(122122391307), 32: new(big.Int).SetUint64(119295484101),
				33: new(big.Int).SetUint64(118730102661), 34: new(big.Int).SetUint64(116694729471),
				35: new(big.Int).SetUint64(113754745977), 36: new(big.Int).SetUint64(21371418479),
				37: new(big.Int).SetUint64(20579884461), 38: new(big.Int).SetUint64(18205282407),
				39: new(big.Int).SetUint64(15830680355), 40: new(big.Int).SetUint64(15491451490),
				41: new(big.Int).SetUint64(15378375202), 42: new(big.Int).SetUint64(14247612319),
				43: new(big.Int).SetUint64(13908383455), 44: new(big.Int).SetUint64(13795307167),
				45: new(big.Int).SetUint64(13116849436), 46: new(big.Int).SetUint64(7915340177),
				47: new(big.Int).SetUint64(7463035025), 48: new(big.Int).SetUint64(6897653583),
				49: new(big.Int).SetUint64(6671501006),
			},
		},
		{
			roundValue: new(big.Int).SetUint64(1750000000000000000),
			minerIdx: map[uint64]uint64{
				1: 847867, 2: 701841, 3: 79392, 4: 22031,
				5: 18602, 6: 16548, 7: 16101, 8: 14813,
				9: 11005, 10: 10942, 11: 10267, 12: 9878,
				13: 9370, 14: 8585, 15: 8519, 16: 8030,
				17: 7400, 18: 7341, 19: 6968, 20: 6711,
				21: 6609, 22: 6368, 23: 6345, 24: 5322,
				25: 5256, 26: 4961, 27: 4815, 28: 4786,
				29: 4483, 30: 4164, 31: 4144, 32: 4124,
				33: 3611, 34: 3435, 35: 3229, 36: 2931,
				37: 2512, 38: 2474, 39: 2378, 40: 2356,
				41: 2341, 42: 2308, 43: 2198, 44: 2196,
				45: 2178, 46: 2169, 47: 2094, 48: 2073,
				49: 1919, 50: 1897, 51: 1865, 52: 1743,
				53: 1735, 54: 1729, 55: 1675, 56: 1651,
				57: 1650, 58: 1648, 59: 1648, 60: 1631,
				61: 1627, 62: 1619, 63: 1613, 64: 1044,
				65: 1028, 66: 953, 67: 947, 68: 873,
				69: 852, 70: 848, 71: 304, 72: 278,
				73: 242, 74: 235, 76: 234, 77: 193,
				78: 125, 79: 124, 80: 68, 81: 58,
				82: 58, 83: 43,
			},
			recipientIdx: map[uint64]uint64{
				500: 9, 501: 31,
				502: 45, 503: 15,
			},
			outputValues: map[uint64]*big.Int{
				// miners
				1: new(big.Int).SetUint64(757633367143662535), 2: new(big.Int).SetUint64(627148078683891763),
				3: new(big.Int).SetUint64(70942763763974368), 4: new(big.Int).SetUint64(19686366743300575),
				5: new(big.Int).SetUint64(16622295590707517), 6: new(big.Int).SetUint64(14786891056608321),
				7: new(big.Int).SetUint64(14387462708632498), 8: new(big.Int).SetUint64(13236537177999701),
				9: new(big.Int).SetUint64(9833800826563607), 10: new(big.Int).SetUint64(9777505556043525),
				11: new(big.Int).SetUint64(9174341943328356), 12: new(big.Int).SetUint64(8826740987259910),
				13: new(big.Int).SetUint64(8372804520209086), 14: new(big.Int).SetUint64(7671347578014408),
				15: new(big.Int).SetUint64(7612371580326703), 16: new(big.Int).SetUint64(7175413052004158),
				17: new(big.Int).SetUint64(6612460346803334), 18: new(big.Int).SetUint64(6559739379173415),
				19: new(big.Int).SetUint64(6226435634665626), 20: new(big.Int).SetUint64(5996786673972591),
				21: new(big.Int).SetUint64(5905641950273410), 22: new(big.Int).SetUint64(5690290201141031),
				23: new(big.Int).SetUint64(5669737959522588), 24: new(big.Int).SetUint64(4755609995363154),
				25: new(big.Int).SetUint64(4696633997675449), 26: new(big.Int).SetUint64(4433029159525857),
				27: new(big.Int).SetUint64(4302567104034872), 28: new(big.Int).SetUint64(4276653408081183),
				29: new(big.Int).SetUint64(4005899964151263), 30: new(big.Int).SetUint64(3720849308660687),
				31: new(big.Int).SetUint64(3702977794209867), 32: new(big.Int).SetUint64(3685106279759047),
				33: new(big.Int).SetUint64(3226701934095519), 34: new(big.Int).SetUint64(3069432606928304),
				35: new(big.Int).SetUint64(2885356008084860), 36: new(big.Int).SetUint64(2619070442767645),
				37: new(big.Int).SetUint64(2244662215022969), 38: new(big.Int).SetUint64(2210706337566412),
				39: new(big.Int).SetUint64(2124923068202476), 40: new(big.Int).SetUint64(2105264402306575),
				41: new(big.Int).SetUint64(2091860766468460), 42: new(big.Int).SetUint64(2062372767624607),
				43: new(big.Int).SetUint64(1964079438145098), 44: new(big.Int).SetUint64(1962292286700016),
				45: new(big.Int).SetUint64(1946207923694278), 46: new(big.Int).SetUint64(1938165742191409),
				47: new(big.Int).SetUint64(1871147563000835), 48: new(big.Int).SetUint64(1852382472827474),
				49: new(big.Int).SetUint64(1714771811556161), 50: new(big.Int).SetUint64(1695113145660260),
				51: new(big.Int).SetUint64(1666518722538948), 52: new(big.Int).SetUint64(1557502484388947),
				53: new(big.Int).SetUint64(1550353878608619), 54: new(big.Int).SetUint64(1544992424273373),
				55: new(big.Int).SetUint64(1496739335256160), 56: new(big.Int).SetUint64(1475293517915176),
				57: new(big.Int).SetUint64(1474399942192635), 58: new(big.Int).SetUint64(1472612790747553),
				59: new(big.Int).SetUint64(1472612790747553), 60: new(big.Int).SetUint64(1457422003464356),
				61: new(big.Int).SetUint64(1453847700574192), 62: new(big.Int).SetUint64(1446699094793864),
				63: new(big.Int).SetUint64(1441337640458618), 64: new(big.Int).SetUint64(932893054332794),
				65: new(big.Int).SetUint64(918595842772138), 66: new(big.Int).SetUint64(851577663581564),
				67: new(big.Int).SetUint64(846216209246318), 68: new(big.Int).SetUint64(780091605778285),
				69: new(big.Int).SetUint64(761326515604924), 70: new(big.Int).SetUint64(757752212714760),
				71: new(big.Int).SetUint64(271647019652461), 72: new(big.Int).SetUint64(248414050866395),
				73: new(big.Int).SetUint64(216245324854919), 74: new(big.Int).SetUint64(209990294797132),
				76: new(big.Int).SetUint64(209096719074591), 77: new(big.Int).SetUint64(172460114450411),
				78: new(big.Int).SetUint64(111696965317623), 79: new(big.Int).SetUint64(110803389595082),
				80: new(big.Int).SetUint64(60763149132787), 81: new(big.Int).SetUint64(51827391907377),
				82: new(big.Int).SetUint64(51827391907377), 83: new(big.Int).SetUint64(38423756069262),
				// recipients
				500: new(big.Int).SetUint64(15750000000005), 501: new(big.Int).SetUint64(54250000000013),
				502: new(big.Int).SetUint64(78750000000019), 503: new(big.Int).SetUint64(26250000000006),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(75770913805747), 2: new(big.Int).SetUint64(62721079976387),
				3: new(big.Int).SetUint64(7094985874985), 4: new(big.Int).SetUint64(1968833557686),
				5: new(big.Int).SetUint64(1662395798650), 6: new(big.Int).SetUint64(1478836989360),
				7: new(big.Int).SetUint64(1438890159879), 8: new(big.Int).SetUint64(1323786096410),
				9: new(big.Int).SetUint64(983478430499), 10: new(big.Int).SetUint64(977848340438),
				11: new(big.Int).SetUint64(917525946927), 12: new(big.Int).SetUint64(882762374963),
				13: new(big.Int).SetUint64(837364188440), 14: new(big.Int).SetUint64(767211478950),
				15: new(big.Int).SetUint64(761313289362), 16: new(big.Int).SetUint64(717613066507),
				17: new(big.Int).SetUint64(661312165897), 18: new(big.Int).SetUint64(656039541872),
				19: new(big.Int).SetUint64(622705834050), 20: new(big.Int).SetUint64(599738641261),
				21: new(big.Int).SetUint64(590623257353), 22: new(big.Int).SetUint64(569085928707),
				23: new(big.Int).SetUint64(567030499002), 24: new(big.Int).SetUint64(475608560393),
				25: new(big.Int).SetUint64(469710370805), 26: new(big.Int).SetUint64(443347250677),
				27: new(big.Int).SetUint64(430299740377), 28: new(big.Int).SetUint64(427708111619),
				29: new(big.Int).SetUint64(400630059421), 30: new(big.Int).SetUint64(372122143080),
				31: new(big.Int).SetUint64(370334812902), 32: new(big.Int).SetUint64(368547482724),
				33: new(big.Int).SetUint64(322702463655), 34: new(big.Int).SetUint64(306973958089),
				35: new(big.Int).SetUint64(288564457254), 36: new(big.Int).SetUint64(261933237600),
				37: new(big.Int).SetUint64(224488670370), 38: new(big.Int).SetUint64(221092743031),
				39: new(big.Int).SetUint64(212513558176), 40: new(big.Int).SetUint64(210547494980),
				41: new(big.Int).SetUint64(209206997346), 42: new(big.Int).SetUint64(206257902553),
				43: new(big.Int).SetUint64(196427586573), 44: new(big.Int).SetUint64(196248853555),
				45: new(big.Int).SetUint64(194640256395), 46: new(big.Int).SetUint64(193835957815),
				47: new(big.Int).SetUint64(187133469647), 48: new(big.Int).SetUint64(185256772960),
				49: new(big.Int).SetUint64(171494330589), 50: new(big.Int).SetUint64(169528267392),
				51: new(big.Int).SetUint64(166668539108), 52: new(big.Int).SetUint64(155765825021),
				53: new(big.Int).SetUint64(155050892950), 54: new(big.Int).SetUint64(154514693897),
				55: new(big.Int).SetUint64(149688902416), 56: new(big.Int).SetUint64(147544106202),
				57: new(big.Int).SetUint64(147454739693), 58: new(big.Int).SetUint64(147276006675),
				59: new(big.Int).SetUint64(147276006675), 60: new(big.Int).SetUint64(145756776024),
				61: new(big.Int).SetUint64(145399309989), 62: new(big.Int).SetUint64(144684377917),
				63: new(big.Int).SetUint64(144148178864), 64: new(big.Int).SetUint64(93298635297),
				65: new(big.Int).SetUint64(91868771155), 66: new(big.Int).SetUint64(85166282987),
				67: new(big.Int).SetUint64(84630083933), 68: new(big.Int).SetUint64(78016962274),
				69: new(big.Int).SetUint64(76140265587), 70: new(big.Int).SetUint64(75782799551),
				71: new(big.Int).SetUint64(27167418707), 72: new(big.Int).SetUint64(24843889476),
				73: new(big.Int).SetUint64(21626695155), 74: new(big.Int).SetUint64(21001129593),
				76: new(big.Int).SetUint64(20911763084), 77: new(big.Int).SetUint64(17247736218),
				78: new(big.Int).SetUint64(11170813614), 79: new(big.Int).SetUint64(11081447105),
				80: new(big.Int).SetUint64(6076922605), 81: new(big.Int).SetUint64(5183257516),
				82: new(big.Int).SetUint64(5183257516), 83: new(big.Int).SetUint64(3842759883),
			},
		},
	}

	for i, tt := range tests {
		outputValues, outputFees, err := CreditRound(tt.roundValue, tt.minerIdx, tt.recipientIdx)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if !reflect.DeepEqual(outputValues, tt.outputValues) {
			t.Errorf("failed on %d: output values mismatch: have %v, want %v", i, outputValues, tt.outputValues)
		} else if !reflect.DeepEqual(outputFees, tt.outputFees) {
			t.Errorf("failed on %d: output fees mismatch: have %v, want %v", i, outputFees, tt.outputFees)
		}
	}
}

func TestProcessFeeBalance(t *testing.T) {
	tests := []struct {
		roundChain  string
		minerChain  string
		value       *big.Int
		fee         *big.Int
		feeBalance  *big.Int
		price       float64
		outputValue *big.Int
		outputFee   *big.Int
	}{
		{
			roundChain:  "ETC",
			minerChain:  "USDC",
			value:       new(big.Int).SetUint64(17750132046676893599),
			fee:         new(big.Int).SetUint64(177501320466768935),
			feeBalance:  new(big.Int),
			price:       0.022,
			outputValue: new(big.Int).SetUint64(454545454545454572),
			outputFee:   new(big.Int).SetUint64(4545454545454545),
		},
		{
			roundChain:  "ETC",
			minerChain:  "USDC",
			value:       new(big.Int).SetUint64(17750132046676893599),
			fee:         new(big.Int).SetUint64(177501320466768935),
			feeBalance:  new(big.Int).SetUint64(5000000000000000),
			price:       0.022,
			outputValue: new(big.Int).SetUint64(227272727272727286),
			outputFee:   new(big.Int).SetUint64(2272727272727272),
		},
		{
			roundChain:  "ETC",
			minerChain:  "USDC",
			value:       new(big.Int).SetUint64(17750132046676893599),
			fee:         new(big.Int).SetUint64(177501320466768935),
			feeBalance:  new(big.Int).SetUint64(9000000000000000),
			price:       0.022,
			outputValue: new(big.Int).SetUint64(45454545454545457),
			outputFee:   new(big.Int).SetUint64(454545454545454),
		},
		{
			roundChain:  "ETC",
			minerChain:  "USDC",
			value:       new(big.Int).SetUint64(17750132046676893599),
			fee:         new(big.Int).SetUint64(177501320466768935),
			feeBalance:  new(big.Int).SetUint64(17750132046676893599),
			price:       0.022,
			outputValue: new(big.Int),
			outputFee:   new(big.Int),
		},
	}

	for i, tt := range tests {
		outputValue, outputFee, err := ProcessFeeBalance(tt.roundChain, tt.minerChain, tt.value, tt.fee, tt.feeBalance, tt.price)
		if err != nil {
			t.Errorf("failed on %d: %v", i, err)
		} else if outputValue.Cmp(tt.outputValue) != 0 {
			t.Errorf("failed on %d: fee balance value mismatch: have %s, want %s", i, outputValue, tt.outputValue)
		} else if outputFee.Cmp(tt.outputFee) != 0 {
			t.Errorf("failed on %d: fee balance fee mismatch: have %s, want %s", i, outputFee, tt.outputFee)
		}
	}
}
