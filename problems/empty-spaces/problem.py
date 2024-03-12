#!/usr/bin/env python3

import os
import typing
from problems import problem_utils

# cat nixos/modules/module-list.nix \
#   | sed -nr 's/^.*\/([a-zA-Z]*)\.nix.*$/\1/p'
#   | sort -u

star_coords = [(1, 9), (1, 10), (2, 9), (2, 10), (3, 8), (3, 9), (3, 10), (3, 11), (4, 8), (4, 9), (4, 10), (4, 11), (5, 7), (5, 8), (5, 9), (5, 10), (5, 11), (5, 12), (6, 7), (6, 8), (6, 9), (6, 10), (6, 11), (6, 12), (7, 7), (7, 8), (7, 9), (7, 10), (7, 11), (7, 12), (8, 1), (8, 2), (8, 3), (8, 4), (8, 5), (8, 6), (8, 7), (8, 8), (8, 9), (8, 10), (8, 11), (8, 12), (8, 13), (8, 14), (8, 15), (8, 16), (8, 17), (8, 18), (9, 2), (9, 3), (9, 4), (9, 5), (9, 6), (9, 7), (9, 8), (9, 9), (9, 10), (9, 11), (9, 12), (9, 13), (9, 14), (9, 15), (9, 16), (9, 17), (10, 4), (10, 5), (10, 6), (10, 7), (10, 8), (10, 9), (10, 10), (10, 11), (10, 12), (10, 13), (10, 14), (10, 15), (11, 5), (11, 6), (11, 7), (11, 8), (11, 9), (11, 10), (11, 11), (11, 12), (11, 13), (11, 14), (12, 5), (12, 6), (12, 7), (12, 8), (12, 9), (12, 10), (12, 11), (12, 12), (12, 13), (12, 14), (13, 4), (13, 5), (13, 6), (13, 7), (13, 8), (13, 9), (13, 10), (13, 11), (13, 12), (13, 13), (13, 14), (13, 15), (14, 4), (14, 5), (14, 6), (14, 7), (14, 8), (14, 9), (14, 10), (14, 11), (14, 12), (14, 13), (14, 14), (14, 15), (15, 3), (15, 4), (15, 5), (15, 6), (15, 7), (15, 12), (15, 13), (15, 14), (15, 15), (15, 16), (16, 3), (16, 4), (16, 5), (16, 13), (16, 14), (16, 15), (16, 16), (17, 2), (17, 3), (17, 4), (17, 15), (17, 16), (17, 17), (18, 2), (18, 17)]
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
