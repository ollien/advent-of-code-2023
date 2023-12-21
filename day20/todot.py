import sys
import typing
import string

def parseModules(f: typing.TextIO) -> dict[str, list[str]]:
    modules = {}
    for line in f:
        (key, lineModules) = line.strip().split(' -> ')
        modules[key] = lineModules.split(', ')

    for fullName in modules:
        for children in modules.values():
            for (i, module) in enumerate(children):
                if fullName == f"&{module}" or fullName == f"%{module}":
                    children[i] = fullName

    return modules

def alpha_only(s: str) -> str:
    return ''.join(char for char in s if char in string.ascii_letters)


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} input.txt")

    with open(sys.argv[1]) as f:
        modules = parseModules(f)

    import pprint
    print("graph {")
    for node in modules.keys():
        print(f'  {alpha_only(node)}[label="{node}"]')

    for parent, children in modules.items():
        print(f"  {alpha_only(parent)} -> {{{' '.join(alpha_only(child) for child in children)}}}")

    print("}")





if __name__ == "__main__":
    main()
