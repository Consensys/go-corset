MINUS_ONE = 0xffffffff

def long_division_u32(a: int, b: int, get_remainder: int = 0) -> int:
    if b == 0:
        if get_remainder == 0:
            return MINUS_ONE  # RISC-V spec: /0 returns -1
        # get_remainder == 1
        return a  # RISC-V spec: /0 returns dividend
    q = 0
    r = 0
    for i in range(32):
        r = ((r << 1) | ((a >> (31 - i)) & 1)) & 0xffffffff
        if r >= b:
            r = r - b
            q = q | (1 << (31 - i))

    if get_remainder == 0:
        return q
    # get_remainder == 1
    return r

long_division_u32(0x00000002, 0x00000001)

test_cases = [
    (0x00000001, 0x00000001),
    (0x00000001, 0x00000000),
    (0x00000000, 0x00000001),
    (0x00000000, 0x00000000),
    (0x00000002, 0x00000001),
    (0x00000004, 0x00000002),
    (0x00000007, 0x00000002),
    (0x00000007, 0x00000003),
    (0x0000000d, 0x00000003),
    (0x000000ff, 0x00000010),
    (0x00000100, 0x00000010),
    (0x00010000, 0x00000002),
    (0x00010000, 0x00010000),
    (0x00010001, 0x00010000),
    (0x7fffffff, 0x00000001),
    (0x7fffffff, 0x00000002),
    (0x7fffffff, 0x7fffffff),
    (0x80000000, 0x00000001),
    (0x80000000, 0x00000002),
    (0x80000000, 0x80000000),
    (0xffffffff, 0x00000001),
    (0xffffffff, 0x00000002),
    (0xffffffff, 0xffffffff),
    (0xffffffff, 0x00000010),
    (0x12345678, 0x00000010),
    (0xdeadbeef, 0x000000ff),
    (0xcafebabe, 0x00000003),
    (0xaaaaaaaa, 0x00000005),
]

for a, b in test_cases:
    q = long_division_u32(a, b, get_remainder=0)
    r = long_division_u32(a, b, get_remainder=1)
    print(f"{r:08x}")
