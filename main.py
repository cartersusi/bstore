import pandas as pd

df = pd.read_csv('time.csv')

# Calculate the average time for each size
result = df.groupby('size')['time'].mean().reset_index()

print("Average time for each size:")
print(result)

# If you want to format the output
result['time'] = result['time'].round(6)
print("\nFormatted output:")
print(result.to_string(index=False))