
def main():
    updtw1 = 10
    max_crime = 16644
    max_enron = 26946
    max_wiki = 9738
    storage2 = updtw1 * 32
    storage3 = updtw1 * 18
    Crime_file_path = 'scripts/search-tokens/data/w2_odxt_crime_storage_data.csv'
    Enron_file_path = 'scripts/search-tokens/data/w2_odxt_enron_storage_data.csv'
    Wiki_file_path = 'scripts/search-tokens/data/w2_odxt_wiki_storage_data.csv'
    with open(Crime_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_crime + 1):
            updtw2 = max_crime
            storage2 = updtw1 * 32
            storage3 = updtw1 * 18
            storage = storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Enron_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_enron + 1):
            updtw2 = max_enron
            storage2 = updtw1 * 32
            storage3 = updtw1 * 18
            storage = storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Wiki_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_wiki + 1):
            updtw2 = max_wiki
            storage2 = updtw1 * 32
            storage3 = updtw1 * 18
            storage = storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")



if __name__ == "__main__":
    main()