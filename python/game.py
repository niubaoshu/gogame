import random

print('''收益分配：
         你选A我选A，我给你7
         你选B我选B，我给你4
         你选A我选B，你给我5
         你选B我选A，你给我6''')


def sal(your_c, my_c, your_s, my_s):
    """

    :param your_c:
    :param my_c:
    :param your_s:
    :param my_s:
    :return:
    """
    if (your_c == 'A' or your_c == 'a') and my_c == 'A':
        your_s += 7
        my_s -= 7
    if (your_c == 'B' or your_c == 'b') and my_c == 'B':
        your_s += 4
        my_s -= 4
    if (your_c == 'A' or your_c == 'a') and my_c == 'B':
        your_s -= 5
        my_s += 5
    if (your_c == 'B' or your_c == 'b') and my_c == 'A':
        your_s -= 6
        my_s += 6
    # your_s += 0.1
    # my_s -= 0.1

    return your_s, my_s


def strategy(a, b):
    """

    :param a:
    :param b:
    :return:
    """
    if random.randint(0, a - 1) < b:
        return 'A'
    else:
        return 'B'


my_sum = 10000
your_sum = 10000
start_sum = your_sum
n = 0
new_game = True
your_strategy = 50
auto = False
my_choice = 'A'
auto_a = 0.0
auto_b = 0.0
my_a = 0.0
my_b = 0.0
while True:
    if new_game:
        my_choice = strategy(22, 9)
    print(
        '------------------------------------------------------------------------------------------------------------')
    print("当前第%d回合，当前积分，你的：%5.1f,我的：%5.1f" % (n, your_sum, my_sum))
    your_choice = input("我已选好，请输入你的选择A或B，如要程序自动帮助选择策略请输入S")
    if your_choice == 's' or your_choice == 'S':
        your_strategy = float(input("请输入选A的概率（0-10000），输入后将按照选A的概率自动帮你选择，直至有一方输光"))
        if your_strategy < 0 or your_strategy > 10000:
            continue
        else:
            print("将按照选A的概率%4.2f%%自动帮你选择" % (your_strategy / 100))
            while your_sum > 0 and my_sum > 0:
                pms = my_sum
                pys = your_sum
                print("当前第%d回合，当前积分，你的：%5.1f,我的：%5.1f" % (n, your_sum, my_sum))
                your_choice = strategy(10000, your_strategy)
                if your_choice == 'A':
                    auto_a += 1
                else:
                    auto_b += 1
                my_choice = strategy(22, 9)
                if my_choice == 'A':
                    my_a += 1
                else:
                    my_b += 1
                (your_sum, my_sum) = sal(your_choice, my_choice, your_sum, my_sum)
                print("你选了%s，我选择了%s，本回合我得分%5.1f,你得分%5.1f" % (your_choice, my_choice, my_sum - pms, your_sum - pys))
                n += 1

            print("my_a%d,my_b%d,实占比%4.2f,应占比%4.2f" % (my_a, my_b, my_a / (my_a + my_b) * 100, 9 / 22 * 100))
            print("游戏结束了，总共%d回合，其中帮你选A %d次，选B %d次，选A的比例%4.2f%%，你输入的策略是%4.2f%%。\n当前积分，你的：%5.1f,我的：%5.1f" % (
                n, auto_a, auto_b, auto_a / (auto_a + auto_b) * 100, your_strategy / 100, your_sum, my_sum))
            print("平均每回合损失%2.8f" % (start_sum / n))
            break

    pms = my_sum
    pys = your_sum
    if your_choice == 'A' or your_choice == 'a' or your_choice == 'B' or your_choice == 'b':
        (your_sum, my_sum) = sal(your_choice, my_choice, your_sum, my_sum)
        n += 1
        print("你选了%s，我选择了%s，本回合我得分%5.1f,你得分%5.1f" % (your_choice, my_choice, my_sum - pms, your_sum - pys))
        new_game = True
    else:
        new_game = False
        print("请输入A或B，其他字符无效")
