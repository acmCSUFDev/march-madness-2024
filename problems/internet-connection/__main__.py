#!/usr/bin/env python3

from logging import debug
import os
import typing
from problems import problem_utils
from math import sqrt, acos, floor, sin, pi

class Problem(problem_utils.Problem):
    def __init__(self, seed=0) -> None:
        super().__init__(seed)
        self.consts = {
            "GRID_SIZE": (10000, 10000),
            "NUM_PTS": 1000,
            "MIN_RADIUS": 50,
            "MAX_RADIUS": 200,
            "NUM_CIRCLES": 15,
            "MIN_DIST": 10,
        }
        # Generate the input
        self.pts = []
        self.radii = []
        self.dist = lambda p1, p2: ((p1[0] - p2[0])**2 + (p1[1] - p2[1])**2)**0.5
        while len(self.pts) < self.consts["NUM_PTS"]:
            x = self.rand.randint(0, self.consts["GRID_SIZE"][0])
            y = self.rand.randint(0, self.consts["GRID_SIZE"][1])
            valid = True
            for p in self.pts:
                if self.dist((x, y), p) < self.consts["MIN_DIST"]:
                    valid = False
                    break
            if valid:
                self.pts.append((x, y))
                self.radii.append(self.rand.randint(self.consts["MIN_RADIUS"], self.consts["MAX_RADIUS"]))
        
    def generate_input(self, output: typing.IO | None = None):
        for pt, r in zip(self.pts, self.radii):
            print(f"Router located at x={pt[0]}, y={pt[1]} with reach={r}ft", file=output)

    # Return the # of pairs of circles that intersect
    def part1_answer(self):
        ans = 0
        for i in range(len(self.pts)):
            for j in range(i+1, len(self.pts)):
                if self.dist(self.pts[i], self.pts[j]) < self.radii[i] + self.radii[j]:
                    ans += 1
        return ans

    # Return the largest intersection area between any two pairs of circles
    def part2_answer(self):
        def intersectionArea(X1, Y1, R1, X2, Y2, R2):
            Pi = pi
            d = sqrt(((X2 - X1) * (X2 - X1)) + ((Y2 - Y1) * (Y2 - Y1)))
            if (d > R1 + R2):
                ans = 0
            elif (d <= (R1 - R2) and R1 >= R2):
                ans = floor(Pi * R2 * R2)
            elif (d <= (R2 - R1) and R2 >= R1):
                ans = floor(Pi * R1 * R1)
            else:
                alpha = acos(((R1 * R1) + (d * d) - (R2 * R2)) / (2 * R1 * d)) * 2
                beta = acos(((R2 * R2) + (d * d) - (R1 * R1)) / (2 * R2 * d)) * 2
                a1 = (0.5 * beta * R2 * R2 ) - (0.5 * R2 * R2 * sin(beta))
                a2 = (0.5 * alpha * R1 * R1) - (0.5 * R1 * R1 * sin(alpha))
                ans = floor(a1 + a2)
            return ans
        ans = 0
        for i in range(len(self.pts)):
            for j in range(i+1, len(self.pts)):
                ans = max(ans, intersectionArea(self.pts[i][0], self.pts[i][1], self.radii[i], self.pts[j][0], self.pts[j][1], self.radii[j]))
        return ans

if __name__ == "__main__":
    problem_utils.main(Problem)
