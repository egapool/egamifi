<?php

/**
 * 有期限先物と無期限先物の価格乖離についての調査
 * 0326と1225の先物と近物を比べて、PERP（現物準拠）との価格乖離具合を調べた
 *
 * 全体的な傾向として、近物はPERPと同等かそれ以下
 * 先物はPERPと同等かそれ以上になった
 */

$url = 'https://ftx.com/api/markets';
$response = file_get_contents($url);
$json = json_decode($response, true);
$result = array_reverse($json['result']);
$list = [];
// $term = '1225';
$term = '0326';
$base = '';
foreach ($result as $r) {
    if (strpos($r['name'], $term) !== false) {
        $base = trim($r['name'], '-'.$term);
        $list[$base] = [
            $term => $r['last'],
        ];
    } else if ($base.'-PERP' === $r['name']) {
        $list[$base]['PERP'] = $r['last']; 
    }
}

$filter = [
"AMPL",
"BRZ",
"MTA",
"DMG",
"PRIV",
"UNISWAP",
"TOMO",
"DRGN",
"RUNE",
"HNT",
"SNX",
"DEFI",
"CUSDT",
"SXP",
"KNC",
"ALT",
"MID",
"SHIT",
"OKB",
"MATIC",
"DOGE",
"YFI",
"XLM",
"UNI",
"EXCH",
"ETH",
"VET",
"BCH",
"AVAX",
];
foreach ($list as $base => $l) {
    if (!in_array($base, $filter)) {
        continue;
    }
    if (isset($l[$term]) && isset($l['PERP'])) {
        $diff = $l[$term] - $l['PERP'];
        if ($diff > 0) {
            continue;
        }
        $rate = $diff/$l['PERP'] * 100;
        print($base.', '. ($rate)."\n");
    }
}

