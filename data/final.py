import pandas as pd

# Read the updated CSV file
df = pd.read_csv("updated_blockusage.csv")

# Round Find % to the nearest 2 decimal places
df['Find %'] = df['Find %'].round(2)

# Write the updated data to a new CSV file
df.to_csv("final-Issue-Numbers.csv", index=False)
