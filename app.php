<?php
var_dump($argv);

$market = $argv[1];
$start_str = $argv[2];
$period = $argv[3];
$start = new DateTimeImmutable($start_str);
$end = (new DateTime($start_str))->add(new DateInterval('P'.$period.'D'));
$interval = new DateInterval('P1D');
for ($start; $start <= $end; $start = $start->add($interval)) {
    echo './bin/egamifi research trades -m '.$market.'-PERP -o '.$market.'-'.$start->format('md').'.csv --end "2021-'.$start->add($interval)->format('m-d').' 0:00:00" -p 86400 && \ ' . "\n";
}
