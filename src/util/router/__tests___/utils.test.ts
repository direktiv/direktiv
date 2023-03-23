import { describe, expect, test } from "vitest";

import { analyzePath } from "../utils";

describe("analyzePath", () => {
  test("undefined", () => {
    expect(analyzePath()).toEqual({
      path: null,
      segments: [],
    });
  });

  test("empyt string", () => {
    expect(analyzePath("")).toEqual({
      path: null,
      segments: [],
    });
  });

  test("/", () => {
    expect(analyzePath("/")).toEqual({
      path: null,
      segments: [],
    });
  });

  test("some/path", () => {
    expect(analyzePath("some/path")).toEqual({
      path: "some/path",
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
