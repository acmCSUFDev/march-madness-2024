import argparse
import random
import json
from io import StringIO
from abc import ABC, abstractmethod
from typing import IO, Type


class Problem(ABC):
    rand: random.Random

    def __init__(self, seed=0) -> None:
        self.rand = random.Random(seed)

    @abstractmethod
    def generate_input(self, output: IO | None = None) -> None:
        pass

    @abstractmethod
    def part1_answer(self) -> int:
        pass

    @abstractmethod
    def part2_answer(self) -> int:
        pass


def main(ProblemClass: Type[Problem]) -> None:
    parser = argparse.ArgumentParser(description="Generate input and answers")
    parser.add_argument("--seed", type=int, default=0, help="random seed")
    parser.add_argument("--part1", action="store_true", help="print part 1 answer")
    parser.add_argument("--part2", action="store_true", help="print part 2 answer")
    parser.add_argument(
        "--json",
        action="store_true",
        help="print JSON of input and answers",
    )

    args = parser.parse_args()

    problem = ProblemClass(args.seed)

    if args.json:
        input = StringIO()
        problem.generate_input(output=input)
        model = {
            "input": input.getvalue(),
            "part1": problem.part1_answer(),
            "part2": problem.part2_answer(),
        }
        print(json.dumps(model))
        return

    if args.part1:
        print(problem.part1_answer())
        return

    if args.part2:
        print(problem.part2_answer())
        return

    problem.generate_input()
