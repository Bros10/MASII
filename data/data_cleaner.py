import pandas as pd

# Read the CSV file
df = pd.read_csv("blockusage.csv")

# Filter the DataFrame by removing rows with names starting with "[" unless it's "[WebApp]"
filtered_df = df[(~df['name'].str.startswith("[")) | (df['name'] == "[WebApp]")]

# Write the filtered data to a new CSV file
filtered_df.to_csv("filtered_blockusage.csv", index=False)
