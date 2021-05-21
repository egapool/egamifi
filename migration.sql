CREATE TABLE `ohlcs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `exchanger` varchar(255)COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `resolution` int NOT NULL,
  `start_time` datetime NOT NULL,
  `market` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  `open` float(20,3) unsigned NOT NULL,
  `high` float(20,3) unsigned NOT NULL,
  `low` float(20,3) unsigned NOT NULL,
  `close` float(20,3) unsigned NOT NULL,
  `volume` float(20,3) unsigned DEFAULT '0.000',
  PRIMARY KEY (`id`),
  KEY `index` (`exchanger`,`resolution`,`start_time`,`market`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci; 
