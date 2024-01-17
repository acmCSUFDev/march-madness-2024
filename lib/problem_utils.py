import argparse
import random
import json
import time
import logging
import contextlib
from io import StringIO
from abc import ABC, abstractmethod
from typing import IO, Type, Callable


class Problem(ABC):
    rand: random.Random

    def __init__(self, seed=0) -> None:
        self.rand = random.Random(seed)

    def coin_flip(self, p: float) -> bool:
        """
        Returns True with probability p, where p is within [0, 1]
        """
        return self.rand.random() < p

    @abstractmethod
    def generate_input(self, output: IO | None = None) -> None:
        pass

    @abstractmethod
    def part1_answer(self) -> int:
        pass

    @abstractmethod
    def part2_answer(self) -> int:
        pass


@contextlib.contextmanager
def measure(what: str, enabled=True):
    if not enabled:
        yield
        return

    start_time = time.process_time_ns()
    yield
    runtime = time.process_time_ns() - start_time
    logging.debug(f"{what} took {runtime / 1000000}ms")


def run_fast_slow(
    fast: Callable[[], int],
    slow: Callable[[], int],
) -> int:
    with measure("fast solution"):
        a = fast()
    with measure("slow solution"):
        b = slow()
    if a != b:
        raise Exception(f"fast solution {a} != slow solution {b}")
    return a


def main(ProblemClass: Type[Problem]) -> None:
    parser = argparse.ArgumentParser(description="Generate input and answers")
    parser.add_argument("--seed", type=int, default=0, help="random seed")
    parser.add_argument("--debug", action="store_true", help="enable debug logging")
    parser.add_argument("--part1", action="store_true", help="print part 1 answer")
    parser.add_argument("--part2", action="store_true", help="print part 2 answer")
    parser.add_argument(
        "--json",
        action="store_true",
        help="print JSON of input and answers",
    )

    args = parser.parse_args()

    if args.debug:
        logging.basicConfig(level=logging.DEBUG)

    with measure("initialization"):
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
        with measure("part 1 solution"):
            print(problem.part1_answer())
        return

    if args.part2:
        with measure("part 2 solution"):
            print(problem.part2_answer())
        return

    with measure("input generation"):
        problem.generate_input()
