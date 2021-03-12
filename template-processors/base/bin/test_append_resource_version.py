import unittest
from appendResourceVersion import get_files
class TestAppendResourceVersion(unittest.TestCase):
    def test_get_files(self):
        expected_list_of_files=["test.yaml", "test.yml", "test.json", "test1.json"]
        actual_list_of_files = get_files("./test/")
        #Convert lists to sets for unordered comparision
        self.assertEqual(set(expected_list_of_files), set(actual_list_of_files))

if __name__ == '__main__':
    unittest.main()
