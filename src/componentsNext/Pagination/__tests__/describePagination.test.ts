import { describe, expect, test } from "vitest";

import describePagination from "../describePagination";

describe("describePagination", () => {
  describe("it describes a simple 11 page pagination", () => {
    test("page 1 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 1 });
      expect(pagination).toStrictEqual([1, 2, "…", 10, 11]);
    });

    test("page 2 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 2 });
      expect(pagination).toStrictEqual([1, 2, 3, "…", 10, 11]);
    });

    test("page 5 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 5 });
      expect(pagination).toStrictEqual([1, 2, 3, 4, 5, 6, "…", 10, 11]);
    });

    test("page 6 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 6 });
      expect(pagination).toStrictEqual([1, 2, "…", 5, 6, 7, "…", 10, 11]);
    });

    test("page 7 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 7 });
      expect(pagination).toStrictEqual([1, 2, "…", 6, 7, 8, 9, 10, 11]);
    });

    test("page 8 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 8 });
      expect(pagination).toStrictEqual([1, 2, "…", 7, 8, 9, 10, 11]);
    });
    test("page 9 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 9 });
      expect(pagination).toStrictEqual([1, 2, "…", 8, 9, 10, 11]);
    });

    test("page 10 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 10 });
      expect(pagination).toStrictEqual([1, 2, "…", 9, 10, 11]);
    });
    test("page 11 / 11", () => {
      const pagination = describePagination({ pages: 11, currentPage: 11 });
      expect(pagination).toStrictEqual([1, 2, "…", 10, 11]);
    });
  });

  describe("configure neighbours", () => {
    test("by default it is configured have 1 neighbour", () => {
      const defaultPagination = describePagination({
        pages: 10,
        currentPage: 1,
      });

      const oneNeighbourPagination = describePagination({
        pages: 10,
        currentPage: 1,
        neighbours: 1,
      });
      expect(defaultPagination).toStrictEqual(oneNeighbourPagination);
    });

    test("increasing neighbours will eventually get rid of the ellipsis", () => {
      const result1Neighbour = describePagination({
        pages: 10,
        currentPage: 3,
        neighbours: 1,
      });
      expect(result1Neighbour).toStrictEqual([1, 2, 3, 4, "…", 9, 10]);

      const result2Neighbour = describePagination({
        pages: 10,
        currentPage: 3,
        neighbours: 2,
      });
      expect(result2Neighbour).toStrictEqual([1, 2, 3, 4, 5, "…", 8, 9, 10]);

      const result3Neighbours = describePagination({
        pages: 10,
        currentPage: 3,
        neighbours: 3,
      });
      expect(result3Neighbours).toStrictEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
    });

    test("when neighbours is very high, it will not have any ellipsis", () => {
      const result = describePagination({
        pages: 10,
        currentPage: 3,
        neighbours: 99,
      });
      expect(result).toStrictEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
    });
  });

  describe("implausible input", () => {
    test("it will return an empty array when current page is less than 1", () => {
      const pagination0 = describePagination({ pages: 1, currentPage: 0 });
      expect(pagination0).toStrictEqual([]);

      const paginationMinus1 = describePagination({
        pages: 1,
        currentPage: -1,
      });
      expect(paginationMinus1).toStrictEqual([]);
    });

    test("it will return an empty array when page is less than 1", () => {
      const pagination0 = describePagination({ pages: 0, currentPage: 1 });
      expect(pagination0).toStrictEqual([]);

      const paginationMinus1 = describePagination({
        pages: -1,
        currentPage: 1,
      });
      expect(paginationMinus1).toStrictEqual([]);
    });

    test("it will return an empty array when currentPage is bigger than pages", () => {
      const pagination = describePagination({ pages: 1, currentPage: 2 });
      expect(pagination).toStrictEqual([]);
    });

    test("it will return an empty array when neighbours is negative", () => {
      const result = describePagination({
        pages: 10,
        currentPage: 1,
        neighbours: -2,
      });
      expect(result).toStrictEqual([]);
    });
  });

  test("it handles 1 page", () => {
    const pagination = describePagination({ pages: 1, currentPage: 1 });
    expect(pagination).toStrictEqual([1]);
  });
});
