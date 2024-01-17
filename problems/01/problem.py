#!/usr/bin/env python3

import os
import typing
from lib import problem_utils

# cat nixos/modules/module-list.nix \
#   | sed -nr 's/^.*\/([a-zA-Z]*)\.nix.*$/\1/p'
#   | sort -u
ALL_SERVICES_FILE = os.path.join(os.path.dirname(__file__), "all-services.txt")


class Problem(problem_utils.Problem):
    all_services: list[str] = []
    all_choices: list[str] = []

    def __init__(self, seed=0) -> None:
        super().__init__(seed)

        self.all_services = open(ALL_SERVICES_FILE, "r").read().splitlines()
        self.rand.shuffle(self.all_services)

        self.all_choices = [
            "[" + self.rand.choices([" OK ", "STOP"], [10, 1])[0] + "]"
            for _ in range(len(self.all_services))
        ]

    def generate_input(self, output: typing.IO | None = None):
        for i in range(len(self.all_services)):
            print(self.all_choices[i], self.all_services[i], file=output)

    def part1_answer(self):
        return len(list(filter(lambda x: x == "[STOP]", self.all_choices)))

    def part2_answer(self):
        return self.part1_answer() + len(self.all_choices)


if __name__ == "__main__":
    problem_utils.main(Problem)
