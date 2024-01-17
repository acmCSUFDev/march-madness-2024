#!/usr/bin/env python3

import os
import enum
import random
import typing
import logging
import pydantic
import itertools
from lib import problem_utils
from datetime import datetime as DateTime, timedelta as TimeDelta


BUILDING_LIST_FILE = os.path.join(os.path.dirname(__file__), "csuf-buildings.txt")
MIN_TIME = DateTime(2023, 10, 1)
MAX_TIME = DateTime(2023, 12, 30, 23, 59, 59)
MIN_STAY_TIME = TimeDelta(hours=2)
MAX_STAY_TIME = TimeDelta(days=7)
PEOPLE = 1000
ENTRIES = 2000
SUSPECTS = 100


class AccessType(str, enum.Enum):
    ENTER = "->"
    LEAVE = "<-"

    def __str__(self) -> str:
        return self.value


class AccessEntry(pydantic.BaseModel):
    name: str
    building: str
    type: AccessType
    time: DateTime

    def __str__(self) -> str:
        time = self.time.strftime("%Y-%m-%d %H:%M:%S")
        return f"{self.name} {self.type} {self.building} [{time}]"


class Problem(problem_utils.Problem):
    accesses: dict[str, list[tuple[DateTime, DateTime]]] = {}
    suspects: list[str] = []
    access_log: list[AccessEntry] = []
    building_list = open(BUILDING_LIST_FILE, "r").read().splitlines()

    def __init__(self, seed=0) -> None:
        super().__init__(seed)

        for _ in range(PEOPLE):
            while True:
                name = generate_name(self.rand)
                if name not in self.accesses:
                    break
            self.accesses[name] = []

        names = self.names()
        # self.suspects = self.rand.sample(names, k=SUSPECTS)

        for _ in range(ENTRIES // 2):
            name = self.rand.choice(names)

            enter_time = generate_time(self.rand, max_time=MAX_TIME - MAX_STAY_TIME)
            leave_time = generate_time(
                self.rand,
                min_time=enter_time + MIN_STAY_TIME,
                max_time=enter_time + MAX_STAY_TIME,
            )

            enter = AccessEntry(
                name=name,
                building=self.rand.choice(self.building_list),
                type=AccessType.ENTER,
                time=enter_time,
            )
            leave = AccessEntry(
                name=name,
                building=enter.building,
                type=AccessType.LEAVE,
                time=leave_time,
            )

            self.access_log.append(enter)
            self.access_log.append(leave)
            self.accesses[name].append((enter.time, leave.time))

    def names(self) -> list[str]:
        return list(self.accesses.keys())

    def generate_input(self, output: typing.IO | None = None):
        for entry in self.access_log:
            print(entry, file=output)

        # print("", file=output)
        # print(f"suspects: {', '.join(self.suspects)}", file=output)

    def part1_answer(self):
        return len(list(filter(lambda e: e.type == AccessType.LEAVE, self.access_log)))

    def part2_answer(self):
        total = 0
        for a, b in itertools.combinations(self.names(), 2):
            a_times = self.accesses[a]
            b_times = self.accesses[b]
            if has_overlapping_times(a_times, b_times):
                total += 1
        return total


def generate_name(rand: random.Random) -> str:
    # VOWELS = "aeiou"
    # CONSONANTS = "bcdfghjklmnpqrstvwxyz"
    # return "".join(
    #     rand.choices(CONSONANTS, k=1)
    #     + rand.choices(VOWELS, k=1)
    #     + rand.choices(CONSONANTS, k=2)
    # )
    return "".join(rand.choices("abcdefghijklmnopqrstuvwxyz", k=4))


def generate_time(
    rand: random.Random,
    min_time: DateTime | None = None,
    max_time: DateTime | None = None,
) -> DateTime:
    min_time = max(min_time, MIN_TIME) if min_time else MIN_TIME
    max_time = min(max_time, MAX_TIME) if max_time else MAX_TIME

    assert min_time < max_time
    assert min_time >= MIN_TIME
    assert max_time <= MAX_TIME

    return DateTime.fromtimestamp(
        rand.randint(
            round(min_time.timestamp()),
            round(max_time.timestamp()),
        )
    )


def has_overlapping_times(
    a: list[tuple[DateTime, DateTime]],
    b: list[tuple[DateTime, DateTime]],
) -> tuple[int, int] | None:
    for i in range(len(a)):
        a_start, a_end = a[i]
        for j in range(len(b)):
            b_start, b_end = b[j]
            if a_start < b_end and b_start < a_end:
                return (i, j)
    return None


if __name__ == "__main__":
    problem_utils.main(Problem)
