#include <iostream>
#include <vector>

/*defiprompt
INPUTS:
3
1
2
3
OUTPUT:
6
-*-
INPUTS:
4
10
2
3
5
OUTPUT:
20
-*-
INPUTS:
5
8
-3
4
0
11
OUTPUT:
20
*/

int main() {
    // Reads the amount of inputs, sums them, and prints the result.
    int count;
    if (!(std::cin >> count)) {
        return 1;
    }

    std::vector<long long> values;
    values.reserve(static_cast<std::size_t>(count));

    long long sum = 0;
    for (int i = 0; i < count; ++i) {
        long long value = 0;
        if (!(std::cin >> value)) {
            return 1;
        }
        values.push_back(value);
        sum += value;
    }

    std::cout << sum << std::endl;
    return 0;
}
