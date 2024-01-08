import pandas as pd

# Read the filtered CSV file
df = pd.read_csv("filtered_blockusage.csv")

# Recalculate Total Rounds
total_rounds = df['count'].sum()
print (total_rounds)

# Recalculate Find % based on the updated Total Rounds
df['Find %'] = (df['count'] / total_rounds) * 100

# Write the updated data to a new CSV file
df.to_csv("updated_blockusage.csv", index=False)
