# Example Python program to plot a complex bar chart 

import pandas as pd

import matplotlib.pyplot as plot
from matplotlib import pyplot as plt
#plotdata['pies'].plot(kind="bar", title="test")
#plt.title("Mince Pie Consumption Study Results")

 

# A python dictionary

data = {"Register":[0.96, 1.01, 0.97, 2.01],

        "Launch":[0.93, 0.91, 0.98, 1.11],
        
        "Buy":[0.87, 0.89, 0.88, 0.90]

        };

index     = ["1000", "3500", "7500", "10000"];

 

# Dictionary loaded into a DataFrame       

dataFrame = pd.DataFrame(data=data, index=index);

 

# Draw a vertical bar chart

ax = dataFrame.plot.bar(rot=360);
ax.set_xlabel("Number of Transactions")
ax.set_ylabel("Average Latency (sec)")
'''
for p in ax.patches:
    ax.annotate(str(p.get_height()), (p.get_x() * 1.005, p.get_height() * 1.005))
'''
plot.show(block=True);
