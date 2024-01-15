import { describe, expect, test } from "vitest";

import { prettifyJsonString } from "../helpers";

describe("prettifyJsonString", () => {
  test("empty json string", () => {
    expect(prettifyJsonString("{}")).toMatchInlineSnapshot('"{}"');
  });

  test("unformatted json", () => {
    expect(prettifyJsonString('{  "some": "json", "multiple":      "keys" }'))
      .toMatchInlineSnapshot(`
      "{
          \\"some\\": \\"json\\",
          \\"multiple\\": \\"keys\\"
      }"
    `);
  });

  test("invalid json", () => {
    expect(prettifyJsonString("")).toMatchInlineSnapshot('"{}"');
    expect(prettifyJsonString("no json")).toMatchInlineSnapshot('"{}"');
    expect(prettifyJsonString('{"some": "invalidJson')).toMatchInlineSnapshot(
      '"{}"'
    );
  });
});
