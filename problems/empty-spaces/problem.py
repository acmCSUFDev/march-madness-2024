#!/usr/bin/env python3

import os
import typing
from problems import problem_utils

# cat nixos/modules/module-list.nix \
#   | sed -nr 's/^.*\/([a-zA-Z]*)\.nix.*$/\1/p'
#   | sort -u


class Problem(problem_utils.Problem):
    def __init__(self, seed=0) -> None:
        super().__init__(seed)
        self.grid = None
        self.consts = {
            "GRID_SIZE" : (300, 200),
            "NUM_PTS" : 70,
            "MIN_PT_DIST" : 15,
            "RECT_MAX_DIM" : 20,
            "CIRC_MAX_RAD": 10
        }
        self.isin = lambda x, y: 0 <= x < self.consts["GRID_SIZE"][0] and 0 <= y < self.consts["GRID_SIZE"][1]
        self.dist = lambda x1, y1, x2, y2: ((x1 - x2) ** 2 + (y1 - y2) ** 2) ** 0.5

    def draw_shape(self, name: str, pt: tuple[int]) -> None:
        match name:
            case "circle":
                radius = self.random.randint(1, self.consts["CIRC_MAX_RAD"])
                for i in range(-radius, radius + 1):
                    for j in range(-radius, radius + 1):
                        if self.isin(pt[0] + i, pt[1] + j) and i ** 2 + j**2 <= radius**2:
                            self.grid[pt[0] + i][pt[1] + j] = "#"
            case "rectangle":
                length, width = self.random.randint(2, self.consts["RECT_MAX_DIM"]), self.random.randint(5, self.consts["RECT_MAX_DIM"])
                if self.random.randint(0,1):
                    length, width = width, length
                x, y = pt[0] - length // 2, pt[1] - width // 2
                for i in range(length):
                    for j in range(width):
                        if self.isin(x + i, y + j):
                            self.grid[x + i][y + j] = "#"
            case "star":
                    global star_coords
                    for x, y in star_coords:
                        if self.isin(pt[0] + x - 10, pt[1] + y - 10):
                            self.grid[pt[0] + x - 10][pt[1] + y - 10] = "#"
            case _:
                raise ValueError(f"Invalid shape name: {name}")

    def generate_input(self, output: typing.IO | None = None):
        # Create the grid
        self.grid = [["." for _ in range(self.consts["GRID_SIZE"][1])] for _ in range(self.consts["GRID_SIZE"][0])]
        # Generate the points
        pts = []
        while len(pts) < self.consts["NUM_PTS"]:
            x, y = self.random.randint(0, self.consts["GRID_SIZE"][0] - 1), self.random.randint(0, self.consts["GRID_SIZE"][1] - 1)
            valid = True
            for xx, yy in pts:
                if self.dist(x, y, xx, yy) < self.consts["MIN_PT_DIST"]:
                    valid = False
                    break
            if valid:
                pts.append((x, y))
                self.grid[x][y] = "x"
        # Draw the shapes
        for pt in pts:
            shape = self.random.choice(["circle", "rectangle", "star"])
            self.draw_shape(shape, pt)
        # Write to output file
        for r in self.grid:
            print(''.join(r), ''.join(r), file=output)

    # Return the number of empty '.' in the grid
    def part1_answer(self):
        return sum([row.count(".") for row in self.grid])

    # Return the number of spots in the grid where we can fit a 15x15 square
    def part2_answer(self):
        ans = 0
        for i in range(self.consts["GRID_SIZE"][0] - 15):
            for j in range(self.consts["GRID_SIZE"][1] - 15):
                valid = True
                for x in range(15):
                    for y in range(15):
                        if self.grid[i + x][j + y] != ".":
                            valid = False
                            break
                    if not valid:
                        break
                if valid:
                    ans += 1
        return ans

if __name__ == "__main__":
    problem_utils.main(Problem)
