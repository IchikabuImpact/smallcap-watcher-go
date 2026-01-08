import unittest
from src.api import parse_data, parse_val

class TestParser(unittest.TestCase):
    def test_parse_val(self):
        self.assertEqual(parse_val("2,493.0円"), 2493.0)
        self.assertEqual(parse_val("13.5倍"), 13.5)
        self.assertEqual(parse_val("24,209,200"), 24209200)
        # simplistic check for large numbers if we were converting them, but for VARCHAR fields we might not use this extensively yet except for calc fields
        self.assertEqual(parse_val("1.36倍"), 1.36)

    def test_parse_data(self):
        raw = {
            "ticker": "8306",
            "companyName": "8306　三菱ＵＦＪ",
            "currentPrice": "2,493.0円",
            "previousClose": "2,496.5 (12/29)",
            "dividendYield": "2.97％",
            "per": "13.5倍",
            "pbr": "1.36倍",
            "marketCap": "29兆5,862億円",
            "volume": "24,209,200"
        }
        
        parsed = parse_data(raw)
        
        self.assertEqual(parsed['ticker'], "8306")
        self.assertEqual(parsed['currentPrice'], 2493.0)
        self.assertEqual(parsed['previousCloseVal'], 2496.5)
        self.assertEqual(parsed['volume'], 24209200)
        self.assertEqual(parsed['pbr'], 1.36)
        # Check that we kept the string fields as requested by DB schema for display
        self.assertEqual(parsed['marketCap'], "29兆5,862億円")

if __name__ == '__main__':
    unittest.main()
