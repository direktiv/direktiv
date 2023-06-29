import { describe, expect, test } from "vitest";

import { generatePaginationPages } from "..";

describe("generatePaginationPages", () => {
  test("page 1 / 10", async () => {
    const results = generatePaginationPages({ pages: 10, currentPage: 1 });
    expect(results).toStrictEqual([1, 2, "...", 9, 10]);
  });

  test("page 2 / 10", async () => {
    const results = generatePaginationPages({ pages: 10, currentPage: 2 });
    expect(results).toStrictEqual([1, 2, 3, "...", 9, 10]);
  });

  test("page 5 / 10", async () => {
    const results = generatePaginationPages({ pages: 10, currentPage: 5 });
    expect(results).toStrictEqual([1, 2, "...", 4, 5, 6, "...", 9, 10]);
  });
});
