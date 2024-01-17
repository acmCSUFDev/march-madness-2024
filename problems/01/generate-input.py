#!/usr/bin/env python3

import argparse
import os
import random
import typing

# cat nixos/modules/module-list.nix \
#   | sed -nr 's/^.*\/([a-zA-Z]*)\.nix.*$/\1/p'
#   | sort -u
ALL_SERVICES_FILE = os.path.join(os.path.dirname(__file__), "all-services.txt")


def generate_input(seed=0, output: typing.IO | None = None):
    rand = random.Random(seed)

    all_services = open(ALL_SERVICES_FILE, "r").read().splitlines()
    rand.shuffle(all_services)

    for service in all_services:
        prefix = "[" + rand.choices([" OK ", "STOP"], [10, 1])[0] + "]"
        print(prefix, service, file=output)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed", type=int, default=0)

    args = parser.parse_args()
    generate_input(seed=args.seed)
