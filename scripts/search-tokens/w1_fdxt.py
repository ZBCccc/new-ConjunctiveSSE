
def main():
    updtw1 = 10
    max_crime = 16644
    max_enron = 26946
    max_wiki = 9738
    storage1 = 0
    storage2 = updtw1 * 32
    storage3 = updtw1 * 18
    Crime_file_path = 'scripts/search-tokens/data/w1_fdxt_crime_storage_data.csv'
    Enron_file_path = 'scripts/search-tokens/data/w1_fdxt_enron_storage_data.csv'
    Wiki_file_path = 'scripts/search-tokens/data/w1_fdxt_wiki_storage_data.csv'
    with open(Crime_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw2 in range(10, max_crime + 1):
            storage1 = (updtw1 + updtw2) * 0.1 * 64
            storage = storage1 + storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Enron_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw2 in range(10, max_enron + 1):
            storage1 = (updtw1 + updtw2) * 0.1 * 64
            storage = storage1 + storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")
    with open(Wiki_file_path, 'w', encoding='utf-8') as csv_file:
        csv_file.write("updtw1,updtw2,storage\n")
        for updtw2 in range(10, max_wiki + 1):
            storage1 = (updtw1 + updtw2) * 0.1 * 64
            storage = storage1 + storage2 + storage3
            csv_file.write(f"{updtw1},{updtw2},{storage}\n")



if __name__ == "__main__":
    main()