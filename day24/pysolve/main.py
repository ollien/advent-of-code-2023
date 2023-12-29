import sys
import json
import itertools

import pydantic
import sympy

from typing import Iterable, TextIO


class Triplet(pydantic.BaseModel):
    x: int
    y: int
    z: int


class Hailstone(pydantic.BaseModel):
    position: Triplet
    velocity: Triplet


def parseInputFile(inputFile: TextIO) -> list[Hailstone]:
    inputData = json.load(inputFile)
    return [Hailstone(**datum) for datum in inputData]


def part1(hailstones: Iterable[Hailstone]) -> int:
    total = 0
    for hailstone1, hailstone2 in itertools.combinations(hailstones, 2):
        t = sympy.var("t")  # type: ignore
        x, y = sympy.var("x y")  # type: ignore
        hailstone1XEq = sympy.Eq(
            hailstone1.position.x + hailstone1.velocity.x * t, x  # type: ignore
        )

        hailstone1YEq = sympy.Eq(
            hailstone1.position.y + hailstone1.velocity.y * t, y  # type: ignore
        )

        hailstone2XEq = sympy.Eq(
            hailstone2.position.x + hailstone2.velocity.x * t, x  # type: ignore
        )

        hailstone2YEq = sympy.Eq(
            hailstone2.position.y + hailstone2.velocity.y * t, y  # type: ignore
        )

        # These expressions only have one solution, and we can guarantee that by their algebraic form
        hailstone1Eq = hailstone1XEq.subs({t: sympy.solve(hailstone1YEq, t)[0]})
        hailstone2Eq = hailstone2XEq.subs({t: sympy.solve(hailstone2YEq, t)[0]})

        solns = sympy.solve(
            [
                hailstone1Eq,
                hailstone2Eq,
            ],
            x,
            y,
            dict=True,
        )
        if not solns:
            continue
        elif len(solns) != 1:
            # Should never happen if they're linear
            raise ValueError(
                f"More than one solution for hailstones {hailstone1} {hailstone2}"
            )

        xSoln = solns[0][x]
        ySoln = solns[0][y]
        halestone1Time = sympy.solve(hailstone1XEq)[0][t].subs({x: xSoln})
        halestone2Time = sympy.solve(hailstone2XEq)[0][t].subs({x: xSoln})
        if (
            halestone1Time < 0
            or halestone2Time < 0
            or not (200000000000000 <= xSoln <= 400000000000000)
            or not (200000000000000 <= ySoln <= 400000000000000)
        ):
            continue

        print(total)
        total += 1

    return total


def main():
    if len(sys.argv) != 2:
        print(f"Usage: {sys.argv[0]} inputFile", sys.stderr)
        sys.exit(1)

    with open(sys.argv[1]) as inputFile:
        inputData = parseInputFile(inputFile)

    print(part1(inputData))


if __name__ == "__main__":
    main()
