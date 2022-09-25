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
				1: new(big.Int).SetUint64(7189766763848396), 2: new(big.Int).SetUint64(1965682233236151603),
				// recipients
				500: new(big.Int).SetUint64(9964000000000001), 501: new(big.Int).SetUint64(9964000000000000),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(72623906705539), 2: new(big.Int).SetUint64(19855376093294461),
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
				1: new(big.Int).SetUint64(87466507523), 2: new(big.Int).SetUint64(28106367521),
				3: new(big.Int).SetUint64(17247614166), 4: new(big.Int).SetUint64(13461653130),
				5: new(big.Int).SetUint64(13140725076), 6: new(big.Int).SetUint64(11303432588),
				7: new(big.Int).SetUint64(10769652507), 8: new(big.Int).SetUint64(10762227436),
				9: new(big.Int).SetUint64(9268138215), 10: new(big.Int).SetUint64(9184812422),
				11: new(big.Int).SetUint64(5981306931), 12: new(big.Int).SetUint64(5077923332),
				13: new(big.Int).SetUint64(4740495120), 14: new(big.Int).SetUint64(3745535650),
				15: new(big.Int).SetUint64(3731510517), 16: new(big.Int).SetUint64(3391607281),
				17: new(big.Int).SetUint64(3253830970), 18: new(big.Int).SetUint64(3216705617),
				19: new(big.Int).SetUint64(2239071311), 20: new(big.Int).SetUint64(1413238451),
				// recipients
				500: new(big.Int).SetUint64(2500023807),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(883500076), 2: new(big.Int).SetUint64(283902703),
				3: new(big.Int).SetUint64(174218325), 4: new(big.Int).SetUint64(135976294),
				5: new(big.Int).SetUint64(132734596), 6: new(big.Int).SetUint64(114176087),
				7: new(big.Int).SetUint64(108784369), 8: new(big.Int).SetUint64(108709369),
				9: new(big.Int).SetUint64(93617558), 10: new(big.Int).SetUint64(92775883),
				11: new(big.Int).SetUint64(60417241), 12: new(big.Int).SetUint64(51292155),
				13: new(big.Int).SetUint64(47883790), 14: new(big.Int).SetUint64(37833694),
				15: new(big.Int).SetUint64(37692025), 16: new(big.Int).SetUint64(34258660),
				17: new(big.Int).SetUint64(32866980), 18: new(big.Int).SetUint64(32491976),
				19: new(big.Int).SetUint64(22616882), 20: new(big.Int).SetUint64(14275136),
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
				1: new(big.Int).SetUint64(956474882728166097), 2: new(big.Int).SetUint64(778700910627193978),
				3: new(big.Int).SetUint64(87571626127444812), 4: new(big.Int).SetUint64(24774664218575186),
				5: new(big.Int).SetUint64(21257335811600208), 6: new(big.Int).SetUint64(18628854876025018),
				7: new(big.Int).SetUint64(17728812852088709), 8: new(big.Int).SetUint64(16696675108221450),
				9: new(big.Int).SetUint64(12470731525684676), 10: new(big.Int).SetUint64(12068847089623563),
				11: new(big.Int).SetUint64(7105182494930039), 12: new(big.Int).SetUint64(6965250588223524),
				13: new(big.Int).SetUint64(6101031132404083), 14: new(big.Int).SetUint64(5881617902688266),
				15: new(big.Int).SetUint64(5554736968621846), 16: new(big.Int).SetUint64(5383460314813070),
				17: new(big.Int).SetUint64(3848687162056007), 18: new(big.Int).SetUint64(3639349029623059),
				19: new(big.Int).SetUint64(3342693387405246), 20: new(big.Int).SetUint64(2838938523261789),
				21: new(big.Int).SetUint64(2684453698257796), 22: new(big.Int).SetUint64(2637436577604407),
				23: new(big.Int).SetUint64(2630719846082494), 24: new(big.Int).SetUint64(1502308950401152),
				25: new(big.Int).SetUint64(1442977821957589), 26: new(big.Int).SetUint64(1410513619601677),
				27: new(big.Int).SetUint64(1300807004743769), 28: new(big.Int).SetUint64(1273940078656118),
				29: new(big.Int).SetUint64(1230281323763685), 30: new(big.Int).SetUint64(1213489494958903),
				31: new(big.Int).SetUint64(1209011673944295), 32: new(big.Int).SetUint64(1181025292602992),
				33: new(big.Int).SetUint64(1175428016334731), 34: new(big.Int).SetUint64(1155277821768993),
				35: new(big.Int).SetUint64(1126171985174038), 36: new(big.Int).SetUint64(211577042940251),
				37: new(big.Int).SetUint64(203740856164686), 38: new(big.Int).SetUint64(180232295837992),
				39: new(big.Int).SetUint64(156723735511297), 40: new(big.Int).SetUint64(153365369750341),
				41: new(big.Int).SetUint64(152245914496689), 42: new(big.Int).SetUint64(141051361960167),
				43: new(big.Int).SetUint64(137692996199211), 44: new(big.Int).SetUint64(136573540945559),
				45: new(big.Int).SetUint64(129856809423646), 46: new(big.Int).SetUint64(78361867755648),
				47: new(big.Int).SetUint64(73884046741040), 48: new(big.Int).SetUint64(68286770472779),
				49: new(big.Int).SetUint64(66047859965475),
				// recipients
				500: new(big.Int).SetUint64(10212615010821569), 501: new(big.Int).SetUint64(10212615010821568),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(9661362451799657), 2: new(big.Int).SetUint64(7865665763911050),
				3: new(big.Int).SetUint64(884561880075200), 4: new(big.Int).SetUint64(250249133520962),
				5: new(big.Int).SetUint64(214720563753538), 6: new(big.Int).SetUint64(188170251272980),
				7: new(big.Int).SetUint64(179078917697866), 8: new(big.Int).SetUint64(168653283921428),
				9: new(big.Int).SetUint64(125966985107926), 10: new(big.Int).SetUint64(121907546359834),
				11: new(big.Int).SetUint64(71769520150809), 12: new(big.Int).SetUint64(70356066547712),
				13: new(big.Int).SetUint64(61626577094991), 14: new(big.Int).SetUint64(59410281845336),
				15: new(big.Int).SetUint64(56108454228503), 16: new(big.Int).SetUint64(54378387018314),
				17: new(big.Int).SetUint64(38875627899555), 18: new(big.Int).SetUint64(36761101309324),
				19: new(big.Int).SetUint64(33764579670760), 20: new(big.Int).SetUint64(28676146699615),
				21: new(big.Int).SetUint64(27115693921796), 22: new(big.Int).SetUint64(26640773511156),
				23: new(big.Int).SetUint64(26572927738207), 24: new(big.Int).SetUint64(15174837882840),
				25: new(big.Int).SetUint64(14575533555127), 26: new(big.Int).SetUint64(14247612319209),
				27: new(big.Int).SetUint64(13139464694382), 28: new(big.Int).SetUint64(12868081602587),
				29: new(big.Int).SetUint64(12427084078421), 30: new(big.Int).SetUint64(12257469646050),
				31: new(big.Int).SetUint64(12212239130750), 32: new(big.Int).SetUint64(11929548410131),
				33: new(big.Int).SetUint64(11873010266008), 34: new(big.Int).SetUint64(11669472947161),
				35: new(big.Int).SetUint64(11375474597717), 36: new(big.Int).SetUint64(2137141847882),
				37: new(big.Int).SetUint64(2057988446108), 38: new(big.Int).SetUint64(1820528240787),
				39: new(big.Int).SetUint64(1583068035468), 40: new(big.Int).SetUint64(1549145148993),
				41: new(big.Int).SetUint64(1537837520168), 42: new(big.Int).SetUint64(1424761231921),
				43: new(big.Int).SetUint64(1390838345447), 44: new(big.Int).SetUint64(1379530716622),
				45: new(big.Int).SetUint64(1311684943673), 46: new(big.Int).SetUint64(791534017734),
				47: new(big.Int).SetUint64(746303502435), 48: new(big.Int).SetUint64(689765358311),
				49: new(big.Int).SetUint64(667150100661),
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
				1: new(big.Int).SetUint64(750132046676893599), 2: new(big.Int).SetUint64(620938691766229468),
				3: new(big.Int).SetUint64(70240360162350860), 4: new(big.Int).SetUint64(19491452221089679),
				5: new(big.Int).SetUint64(16457718406641106), 6: new(big.Int).SetUint64(14640486194661704),
				7: new(big.Int).SetUint64(14245012582804453), 8: new(big.Int).SetUint64(13105482354455150),
				9: new(big.Int).SetUint64(9736436461944165), 10: new(big.Int).SetUint64(9680698570340123),
				11: new(big.Int).SetUint64(9083506874582530), 12: new(big.Int).SetUint64(8739347512138525),
				13: new(big.Int).SetUint64(8289905465553551), 14: new(big.Int).SetUint64(7595393641598424),
				15: new(big.Int).SetUint64(7537001564679904), 16: new(big.Int).SetUint64(7104369358419959),
				17: new(big.Int).SetUint64(6546990442379539), 18: new(big.Int).SetUint64(6494791464528134),
				19: new(big.Int).SetUint64(6164787757094679), 20: new(big.Int).SetUint64(5937412548487714),
				21: new(big.Int).SetUint64(5847170247795455), 22: new(big.Int).SetUint64(5633950694199041),
				23: new(big.Int).SetUint64(5613601940121375), 24: new(big.Int).SetUint64(4708524747884311),
				25: new(big.Int).SetUint64(4650132670965791), 26: new(big.Int).SetUint64(4389137781708769),
				27: new(big.Int).SetUint64(4259967429737497), 28: new(big.Int).SetUint64(4234310305030874),
				29: new(big.Int).SetUint64(3966237588268577), 30: new(big.Int).SetUint64(3684009216495729),
				31: new(big.Int).SetUint64(3666314647732541), 32: new(big.Int).SetUint64(3648620078969353),
				33: new(big.Int).SetUint64(3194754390193583), 34: new(big.Int).SetUint64(3039042185077529),
				35: new(big.Int).SetUint64(2856788126816693), 36: new(big.Int).SetUint64(2593139052245193),
				37: new(big.Int).SetUint64(2222437836656405), 38: new(big.Int).SetUint64(2188818156006348),
				39: new(big.Int).SetUint64(2103884225943046), 40: new(big.Int).SetUint64(2084420200303539),
				41: new(big.Int).SetUint64(2071149273731148), 42: new(big.Int).SetUint64(2041953235271888),
				43: new(big.Int).SetUint64(1944633107074354), 44: new(big.Int).SetUint64(1942863650198036),
				45: new(big.Int).SetUint64(1926938538311167), 46: new(big.Int).SetUint64(1918975982367732),
				47: new(big.Int).SetUint64(1852621349505777), 48: new(big.Int).SetUint64(1834042052304430),
				49: new(big.Int).SetUint64(1697793872827883), 50: new(big.Int).SetUint64(1678329847188376),
				51: new(big.Int).SetUint64(1650018537167275), 52: new(big.Int).SetUint64(1542081667711829),
				53: new(big.Int).SetUint64(1535003840206554), 54: new(big.Int).SetUint64(1529695469577597),
				55: new(big.Int).SetUint64(1481920133916990), 56: new(big.Int).SetUint64(1460686651401164),
				57: new(big.Int).SetUint64(1459801922963005), 58: new(big.Int).SetUint64(1458032466086686),
				59: new(big.Int).SetUint64(1458032466086686), 60: new(big.Int).SetUint64(1442992082637976),
				61: new(big.Int).SetUint64(1439453168885339), 62: new(big.Int).SetUint64(1432375341380064),
				63: new(big.Int).SetUint64(1427066970751107), 64: new(big.Int).SetUint64(923656489438410),
				65: new(big.Int).SetUint64(909500834427860), 66: new(big.Int).SetUint64(843146201565905),
				67: new(big.Int).SetUint64(837837830936949), 68: new(big.Int).SetUint64(772367926513153),
				69: new(big.Int).SetUint64(753788629311806), 70: new(big.Int).SetUint64(750249715559168),
				71: new(big.Int).SetUint64(268957445200456), 72: new(big.Int).SetUint64(245954505808312),
				73: new(big.Int).SetUint64(214104282034574), 74: new(big.Int).SetUint64(207911182967458),
				76: new(big.Int).SetUint64(207026454529298), 77: new(big.Int).SetUint64(170752588564763),
				78: new(big.Int).SetUint64(110591054769924), 79: new(big.Int).SetUint64(109706326331765),
				80: new(big.Int).SetUint64(60161533794839), 81: new(big.Int).SetUint64(51314249413245),
				82: new(big.Int).SetUint64(51314249413245), 83: new(big.Int).SetUint64(38043322840854),
				// recipients
				500: new(big.Int).SetUint64(1575000000000005), 501: new(big.Int).SetUint64(5425000000000011),
				502: new(big.Int).SetUint64(7875000000000017), 503: new(big.Int).SetUint64(2625000000000005),
			},
			outputFees: map[uint64]*big.Int{
				1: new(big.Int).SetUint64(7577091380574683), 2: new(big.Int).SetUint64(6272107997638682),
				3: new(big.Int).SetUint64(709498587498493), 4: new(big.Int).SetUint64(196883355768582),
				5: new(big.Int).SetUint64(166239579865061), 6: new(big.Int).SetUint64(147883698935977),
				7: new(big.Int).SetUint64(143889015987924), 8: new(big.Int).SetUint64(132378609640961),
				9: new(big.Int).SetUint64(98347843049941), 10: new(big.Int).SetUint64(97784834043840),
				11: new(big.Int).SetUint64(91752594692753), 12: new(big.Int).SetUint64(88276237496348),
				13: new(big.Int).SetUint64(83736418843975), 14: new(big.Int).SetUint64(76721147894934),
				15: new(big.Int).SetUint64(76131328936161), 16: new(big.Int).SetUint64(71761306650706),
				17: new(big.Int).SetUint64(66131216589692), 18: new(big.Int).SetUint64(65603954187153),
				19: new(big.Int).SetUint64(62270583404997), 20: new(big.Int).SetUint64(59973864126138),
				21: new(big.Int).SetUint64(59062325735308), 22: new(big.Int).SetUint64(56908592870697),
				23: new(big.Int).SetUint64(56703049900215), 24: new(big.Int).SetUint64(47560856039236),
				25: new(big.Int).SetUint64(46971037080463), 26: new(big.Int).SetUint64(44334725067765),
				27: new(big.Int).SetUint64(43029974037752), 28: new(big.Int).SetUint64(42770811161928),
				29: new(big.Int).SetUint64(40063005942107), 30: new(big.Int).SetUint64(37212214308038),
				31: new(big.Int).SetUint64(37033481290228), 32: new(big.Int).SetUint64(36854748272418),
				33: new(big.Int).SetUint64(32270246365591), 34: new(big.Int).SetUint64(30697395808864),
				35: new(big.Int).SetUint64(28856445725421), 36: new(big.Int).SetUint64(26193323760052),
				37: new(big.Int).SetUint64(22448867036934), 38: new(big.Int).SetUint64(22109274303095),
				39: new(big.Int).SetUint64(21251355817606), 40: new(big.Int).SetUint64(21054749498016),
				41: new(big.Int).SetUint64(20920699734658), 42: new(big.Int).SetUint64(20625790255272),
				43: new(big.Int).SetUint64(19642758657317), 44: new(big.Int).SetUint64(19624885355535),
				45: new(big.Int).SetUint64(19464025639506), 46: new(big.Int).SetUint64(19383595781492),
				47: new(big.Int).SetUint64(18713346964705), 48: new(big.Int).SetUint64(18525677296004),
				49: new(big.Int).SetUint64(17149433058867), 50: new(big.Int).SetUint64(16952826739276),
				51: new(big.Int).SetUint64(16666853910781), 52: new(big.Int).SetUint64(15576582502139),
				53: new(big.Int).SetUint64(15505089295015), 54: new(big.Int).SetUint64(15451469389673),
				55: new(big.Int).SetUint64(14968890241586), 56: new(big.Int).SetUint64(14754410620214),
				57: new(big.Int).SetUint64(14745473969323), 58: new(big.Int).SetUint64(14727600667542),
				59: new(big.Int).SetUint64(14727600667542), 60: new(big.Int).SetUint64(14575677602404),
				61: new(big.Int).SetUint64(14539930998842), 62: new(big.Int).SetUint64(14468437791717),
				63: new(big.Int).SetUint64(14414817886375), 64: new(big.Int).SetUint64(9329863529681),
				65: new(big.Int).SetUint64(9186877115433), 66: new(big.Int).SetUint64(8516628298646),
				67: new(big.Int).SetUint64(8463008393302), 68: new(big.Int).SetUint64(7801696227406),
				69: new(big.Int).SetUint64(7614026558705), 70: new(big.Int).SetUint64(7578279955143),
				71: new(big.Int).SetUint64(2716741870712), 72: new(big.Int).SetUint64(2484388947559),
				73: new(big.Int).SetUint64(2162669515500), 74: new(big.Int).SetUint64(2100112959267),
				76: new(big.Int).SetUint64(2091176308377), 77: new(big.Int).SetUint64(1724773621866),
				78: new(big.Int).SetUint64(1117081361313), 79: new(big.Int).SetUint64(1108144710422),
				80: new(big.Int).SetUint64(607692260553), 81: new(big.Int).SetUint64(518325751648),
				82: new(big.Int).SetUint64(518325751648), 83: new(big.Int).SetUint64(384275988291),
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
