<?php

function fetch($url)
{
    $response = @file_get_contents($url);
    if (!$response) {
        return $response;
    }
    $json = json_decode($response, true);
    return $json['result'];
}

function getRate($market)
{
    $url = 'https://ftx.com/api/markets/'.$market.'/candles?resolution=60&limit=0&end_time=1608865140';
    return fetch($url)[0];
}

$url = 'https://ftx.com/api/markets';
$markets = fetch($url);
foreach($markets as $m) {
    if (strpos($m['name'], '-0326') !== false) {
        $base = $m['underlying'];
        $_1225 = $base . '-1225';
        $perp = $base . '-PERP';

        $i = getRate($_1225);
        if (!$i) {
            continue;
        }
        $j = getRate($perp);
        if (!$j) {
            continue;
        }
        print($base .': '.$i['close'].'. '. $j['close'].', '.(($i['close'] - $j['close'])*2/($i['close']+$j['close'])). "\n"); 

    }
}

