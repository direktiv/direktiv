import { analyzePath, checkHandlerInMatcher } from "../utils";
import { describe, expect, test } from "vitest";

describe("analyzePath", () => {
  test("undefined", () => {
    expect(analyzePath()).toEqual({
      path: null,
      isRoot: true,
      parent: null,
      segments: [],
    });
  });

  test("empyt string", () => {
    expect(analyzePath("")).toEqual({
      path: null,
      isRoot: true,
      parent: null,
      segments: [],
    });
  });

  test("/", () => {
    expect(analyzePath("/")).toEqual({
      path: null,
      isRoot: true,
      parent: null,
      segments: [],
    });
  });

  test("some/path", () => {
    expect(analyzePath("some/path")).toEqual({
      path: "some/path",
      isRoot: false,
      parent: {
        relative: "some",
        absolute: "some",
      },
      segments: [
        {
          relative: "some",
          absolute: "some",
        },
        {
          relative: "path",
          absolute: "some/path",
        },
      ],
    });
  });

  test("some/nested/path", () => {
    expect(analyzePath("some/nested/path")).toEqual({
      path: "some/nested/path",
      isRoot: false,
      parent: {
        relative: "nested",
        absolute: "some/nested",
      },
      segments: [
        {
          relative: "some",
          absolute: "some",
        },
        {
          relative: "nested",
          absolute: "some/nested",
        },
        {
          relative: "path",
          absolute: "some/nested/path",
        },
      ],
    });
  });
});

describe("checkHandlerInMatcher", () => {
  test("basic functionality", () => {
    const matcher = {
      id: "id",
      pathname: "pathname",
      params: {},
      handle: { isHome: true },
      data: {},
    };

    expect(checkHandlerInMatcher(matcher, "isHome")).toBe(true);
    expect(checkHandlerInMatcher(matcher, "isWhatever")).toBe(false);
  });
});
