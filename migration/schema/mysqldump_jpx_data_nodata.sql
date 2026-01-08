-- MySQL dump 10.13  Distrib 8.0.36, for Linux (x86_64)
--
-- Host: localhost    Database: jpx_data
-- ------------------------------------------------------
-- Server version	8.0.36

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `watch_detail`
--

DROP TABLE IF EXISTS `watch_detail`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `watch_detail` (
  `ticker` varchar(10) NOT NULL,
  `yymmdd` date NOT NULL,
  `companyName` varchar(255) DEFAULT NULL,
  `currentPrice` decimal(10,2) DEFAULT NULL,
  `previousClose` varchar(20) DEFAULT NULL,
  `dividendYield` varchar(20) DEFAULT NULL,
  `per` varchar(20) DEFAULT NULL,
  `pbr` decimal(5,2) DEFAULT NULL,
  `marketCap` varchar(50) DEFAULT NULL,
  `volume` int DEFAULT NULL,
  `pricemovement` varchar(50) DEFAULT NULL,
  `signal_val` varchar(50) DEFAULT NULL,
  `memo` text,
  PRIMARY KEY (`ticker`,`yymmdd`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `watch_list`
--

DROP TABLE IF EXISTS `watch_list`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `watch_list` (
  `ticker` varchar(10) NOT NULL,
  `companyName` varchar(255) DEFAULT NULL,
  `currentPrice` decimal(10,2) DEFAULT NULL,
  `previousClose` varchar(20) DEFAULT NULL,
  `dividendYield` varchar(20) DEFAULT NULL,
  `per` varchar(20) DEFAULT NULL,
  `pbr` decimal(5,2) DEFAULT NULL,
  `marketCap` varchar(50) DEFAULT NULL,
  `volume` int DEFAULT NULL,
  `pricemovement` varchar(50) DEFAULT NULL,
  `signal_val` varchar(50) DEFAULT NULL,
  `memo` text,
  PRIMARY KEY (`ticker`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-01-01 00:00:00
