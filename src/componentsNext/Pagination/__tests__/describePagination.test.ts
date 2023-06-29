import { describe, expect, test } from "vitest";

import describePagination from "../describePagination";

describe("describePagination", () => {
  describe("it describes a simple 10 page pagination", () => {
    test("page 1 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 1 });
      expect(pagination).toStrictEqual([1, 2, "...", 9, 10]);
    });

    test("page 2 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 2 });
      expect(pagination).toStrictEqual([1, 2, 3, "...", 9, 10]);
    });

    test("page 5 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 5 });
      expect(pagination).toStrictEqual([1, 2, "...", 4, 5, 6, "...", 9, 10]);
    });

    test("page 6 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 6 });
      expect(pagination).toStrictEqual([1, 2, "...", 5, 6, 7, "...", 9, 10]);
    });

    test("page 7 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 7 });
      expect(pagination).toStrictEqual([1, 2, "...", 6, 7, 8, 9, 10]);
    });

    test("page 8 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 8 });
      expect(pagination).toStrictEqual([1, 2, "...", 7, 8, 9, 10]);
    });
    test("page 9 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 9 });
      expect(pagination).toStrictEqual([1, 2, "...", 8, 9, 10]);
    });

    test("page 10 / 10", () => {
      const pagination = describePagination({ pages: 10, currentPage: 10 });
      expect(pagination).toStrictEqual([1, 2, "...", 9, 10]);
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

    test("increasing neighbours will eventually get rid of the ...", () => {
      const result1Neighbour = describePagination({
        pages: 10,
        currentPage: 1,
        neighbours: 2,
      });
      expect(result1Neighbour).toStrictEqual([1, 2, 3, "...", 8, 9, 10]);

      const result2Neighbours = describePagination({
        pages: 10,
        currentPage: 1,
        neighbours: 3,
      });
      expect(result2Neighbours).toStrictEqual([1, 2, 3, 4, "...", 7, 8, 9, 10]);

      const result3Neighbours = describePagination({
        pages: 10,
        currentPage: 1,
        neighbours: 3,
      });
      expect(result3Neighbours).toStrictEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10]);
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
  });

  test("it handles 1 page", () => {
    const pagination = describePagination({ pages: 1, currentPage: 1 });
    expect(pagination).toStrictEqual([1]);
  });
});
