import { describe, expect, test } from "vitest";

import describePagination from "../describePagination";

describe("describePagination", () => {
  describe("it describes a simple 10 page pagination", () => {
    test("page 1 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 1 });
      expect(results).toStrictEqual([1, 2, "...", 9, 10]);
    });

    test("page 2 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 2 });
      expect(results).toStrictEqual([1, 2, 3, "...", 9, 10]);
    });

    test("page 5 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 5 });
      expect(results).toStrictEqual([1, 2, "...", 4, 5, 6, "...", 9, 10]);
    });

    test("page 6 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 6 });
      expect(results).toStrictEqual([1, 2, "...", 5, 6, 7, "...", 9, 10]);
    });

    test("page 7 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 7 });
      expect(results).toStrictEqual([1, 2, "...", 6, 7, 8, 9, 10]);
    });

    test("page 8 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 8 });
      expect(results).toStrictEqual([1, 2, "...", 7, 8, 9, 10]);
    });
    test("page 9 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 9 });
      expect(results).toStrictEqual([1, 2, "...", 8, 9, 10]);
    });

    test("page 10 / 10", () => {
      const results = describePagination({ pages: 10, currentPage: 10 });
      expect(results).toStrictEqual([1, 2, "...", 9, 10]);
    });
  });

  test("it handles 1 page", () => {
    const results = describePagination({ pages: 1, currentPage: 1 });
    expect(results).toStrictEqual([1]);
  });

  test("it will return an empty array when current page is less than 1", () => {
    const resultsWith0 = describePagination({ pages: 1, currentPage: 0 });
    expect(resultsWith0).toStrictEqual([]);

    const resultsWith1 = describePagination({ pages: 1, currentPage: -1 });
    expect(resultsWith1).toStrictEqual([]);
  });

  test("it will return an empty array when page is less than 1", () => {
    const resultsWith0 = describePagination({ pages: 0, currentPage: 1 });
    expect(resultsWith0).toStrictEqual([]);

    const resultsWith1 = describePagination({ pages: -1, currentPage: 1 });
    expect(resultsWith1).toStrictEqual([]);
  });

  test("it will return an empty array when currentPage is bigger than pages", () => {
    const results = describePagination({ pages: 1, currentPage: 2 });
    expect(results).toStrictEqual([]);
  });
});
