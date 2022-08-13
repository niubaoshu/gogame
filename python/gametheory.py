import nashpy as nash
import numpy as np

# A = np.array([[1, 2, 3], [4, 5, 6], [7, 8, 9]])
# A = np.array([[0, -1, 1], [1, 0, -1], [-1, 1, 0]])
# B = np.ones((3, 3), dtype=int) * 10 - A
# B = np.full((3, 3), 10, dtype=int) - A

A = np.random.randint(0, 30, dtype=int, size=(2, 2))

rps = nash.Game(A, A, A)
print(rps)

sigma_r = [1, 0]
sigma_c = [0, 1]
print(rps[sigma_r, sigma_c])
print(rps.zero_sum)
