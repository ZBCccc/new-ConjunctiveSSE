
def main():
    updtw1 = 10
    max_crime = 16644
    max_enron = 26946
    max_wiki = 9738
    storage1 = updtw1 * 0.1 * 2.4
    storage2 = 32
    storage3 = updtw1 * 32
    storage5 = 32
    storage7 = updtw1 * 18
    Crime_file_path = 'scripts/search-tokens/data/w2_sdssecq_crime_storage_data.csv'
    Enron_file_path = 'scripts/search-tokens/data/w2_sdssecq_enron_storage_data.csv'
    Wiki_file_path = 'scripts/search-tokens/data/w2_sdssecq_wiki_storage_data.csv'
    with open(Crime_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_crime + 1):
            updtw2 = max_crime
            storage1 = updtw1 * 0.1 * 2.4
            storage3 = updtw1 * 32
            storage4 = updtw2 * 0.1 * 2.4
            storage7 = updtw1 * 18
            storage6 = updtw2 * 32
            storage = storage1 + storage2 + storage3 + storage4 + storage5 + storage6 + storage7
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Enron_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_enron + 1):
            updtw2 = max_enron
            storage1 = updtw1 * 0.1 * 2.4
            storage3 = updtw1 * 32
            storage4 = updtw2 * 0.1 * 2.4
            storage6 = updtw2 * 32
            storage7 = updtw1 * 18
            storage = storage1 + storage2 + storage3 + storage4 + storage5 + storage6 + storage7
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Wiki_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw1 in range(1, max_wiki + 1):
            updtw2 = max_wiki
            storage1 = updtw1 * 0.1 * 2.4
            storage3 = updtw1 * 32
            storage4 = updtw2 * 0.1 * 2.4
            storage6 = updtw2 * 32
            storage7 = updtw1 * 18
            storage = storage1 + storage2 + storage3 + storage4 + storage5 + storage6 + storage7
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")



if __name__ == "__main__":
    main()