import { describe, expect, test } from "vitest";

import { deepSortObject } from "../utils";

describe("deepSortObject", () => {
  test("should sort the keys of a simple object", () => {
    const obj = { b: 2, a: 1, c: 3 };
    const expected = { a: 1, b: 2, c: 3 };
    expect(deepSortObject(obj)).toEqual(expected);
  });

  test("should deeply sort nested objects", () => {
    const obj = {
      b: { y: 26, x: 25, z: 27 },
      a: { c: 3, b: 2, a: 1 },
      c: { z: { b: 2, a: 1 }, y: 2 },
    };
    const expected = {
      a: { a: 1, b: 2, c: 3 },
      b: { x: 25, y: 26, z: 27 },
      c: { y: 2, z: { a: 1, b: 2 } },
    };
    expect(deepSortObject(obj)).toEqual(expected);
  });

  test("should sort arrays", () => {
    const arr = [3, 1, 2];
    const expected = [3, 1, 2];
    expect(deepSortObject(arr)).toEqual(expected);
  });

  test("should deeply sort arrays of objects", () => {
    const arr = [
      { b: 2, a: 1 },
      { d: 4, c: 3 },
    ];
    const expected = [
      { a: 1, b: 2 },
      { c: 3, d: 4 },
    ];
    expect(deepSortObject(arr)).toEqual(expected);
  });

  test("should handle empty objects and arrays", () => {
    expect(deepSortObject({})).toEqual({});
    expect(deepSortObject([])).toEqual([]);
  });

  test("should handle a complex nested structure", () => {
    const obj = {
      z: [
        { c: 3, b: 2, a: 1 },
        { f: 6, e: 5, d: 4 },
      ],
      y: { b: [4, 2, 1], a: { d: 8, c: 7 } },
      x: 10,
    };

    const expected = {
      x: 10,
      y: { a: { c: 7, d: 8 }, b: [4, 2, 1] },
      z: [
        { a: 1, b: 2, c: 3 },
        { d: 4, e: 5, f: 6 },
      ],
    };

    expect(deepSortObject(obj)).toEqual(expected);
  });

  describe("with custom compare function", () => {
    const sortSpecialStringToTheTop = (a: string, b: string) => {
      if (a === "I_SHOULD_BE_FIRST") {
        return -1;
      }

      if (b === "I_SHOULD_BE_FIRST") {
        return 1;
      }

      return b.localeCompare(a);
    };

    test("should sort with a custom compare function", () => {
      const obj = { b: 1, I_SHOULD_BE_FIRST: 2, c: 3 };
      const expected = { I_SHOULD_BE_FIRST: 2, b: 1, c: 3 };
      expect(deepSortObject(obj, sortSpecialStringToTheTop)).toEqual(expected);
    });

    test("should deeply sort with a custom compare function", () => {
      const obj = {
        b: { y: 1, x: 2, z: 3 },
        I_SHOULD_BE_FIRST: { c: 1, b: 2, I_SHOULD_BE_FIRST: 3 },
        c: { z: { b: 1, a: 2 }, y: 2 },
      };
      const expected = {
        I_SHOULD_BE_FIRST: { I_SHOULD_BE_FIRST: 3, b: 2, c: 1 },
        b: { x: 2, y: 1, z: 3 },
        c: { y: 2, z: { a: 2, b: 1 } },
      };
      const compare = (a: string, b: string) => b.localeCompare(a);
      expect(deepSortObject(obj, compare)).toEqual(expected);
    });
  });
});
