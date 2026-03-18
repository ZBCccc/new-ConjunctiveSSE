from search_time_fixed_w2_common import plot_all


OUTPUT_DIR = "pic/client_search_time_fixed_w2_figures"

DATASET_CONFIG = {
    "Crime": {
        "file_paths": {
            "Nomos": "pic/client_search_time_fixed_w2/Nomos_Crime.csv",
            "MC-ODXT": "pic/client_search_time_fixed_w2/MC-ODXT_Crime.csv",
            "VQNomos": "pic/client_search_time_fixed_w2/VQNomos_Crime.csv",
        },
        "xlim": 17000,
        "x_ticks": [0, 4000, 8000, 12000, 16000],
        "x_minor": 2000,
        "ylim": (0, 15),
        "y_ticks": [0, 3, 6, 9, 12, 15],
        "y_labels": ["", "3", "6", "9", "12", "15"],
        "y_minor": 1.5,
        "legend_xlim": 16644,
        "sample_segments": [(0.15, 12), (0.50, 10), (1.00, 8)],
        "output_name": "Crime",
    },
    "Enron": {
        "file_paths": {
            "Nomos": "pic/client_search_time_fixed_w2/Nomos_Enron.csv",
            "MC-ODXT": "pic/client_search_time_fixed_w2/MC-ODXT_Enron.csv",
            "VQNomos": "pic/client_search_time_fixed_w2/VQNomos_Enron.csv",
        },
        "xlim": 28000,
        "x_ticks": [0, 5000, 10000, 15000, 20000, 25000],
        "x_minor": 2500,
        "ylim": (0, 24),
        "y_ticks": [0, 6, 12, 18, 24],
        "y_labels": ["", "6", "12", "18", "24"],
        "y_minor": 3,
        "legend_xlim": 26946,
        "sample_segments": [(0.12, 12), (0.45, 10), (1.00, 8)],
        "output_name": "Enron",
    },
    "Wikipedia": {
        "file_paths": {
            "Nomos": "pic/client_search_time_fixed_w2/Nomos_Wiki.csv",
            "MC-ODXT": "pic/client_search_time_fixed_w2/MC-ODXT_Wiki.csv",
            "VQNomos": "pic/client_search_time_fixed_w2/VQNomos_Wiki.csv",
        },
        "xlim": 10000,
        "x_ticks": [0, 2000, 4000, 6000, 8000],
        "x_minor": 1000,
        "ylim": (0, 9),
        "y_ticks": [0, 2, 4, 6, 8],
        "y_labels": ["", "2", "4", "6", "8"],
        "y_minor": 1,
        "legend_xlim": 9738,
        "sample_segments": [(0.20, 12), (0.60, 10), (1.00, 8)],
        "output_name": "Wikipedia",
    },
}


plot_all(
    output_dir=OUTPUT_DIR,
    dataset_config=DATASET_CONFIG,
    y_column="client_time_ms",
    y_label="Client Search Time (s)",
    output_suffix="client_search_time_comparison",
)
