import pandas as pd

# Read the CSV file
df = pd.read_csv("filtered_blockusage_without_nonweb.csv")

# Remove rows with all values as NaN
df = df.dropna(how="all")

# Remove rows where 'count' is 4 or below
df = df[df['count'] > 4]

# Recalculate the total rounds (sum of 'count' column)
total_rounds = df['count'].sum()
print(f"Total Rounds: {total_rounds}")

# Recalculate Find % for each row based on updated total rounds
df['Find %'] = ((df['count'] / total_rounds) * 100).round(2)

# Save the updated DataFrame to a new CSV file
df.to_csv("updated_filtered_blockusage.csv", index=False)
