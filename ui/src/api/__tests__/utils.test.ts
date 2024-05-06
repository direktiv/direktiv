import { describe, expect, test } from "vitest";

import { buildSearchParamsString } from "../utils";

describe("buildSearchParamsString", () => {
  test("should return a query string from the provided object", () => {
    expect(
      buildSearchParamsString({
        param1: "value1",
        param2: 2,
        thisWillBeSkipped: undefined,
      })
    ).toEqual("?param1=value1&param2=2");
  });

  test("should return query string without question mark when withoutQuestionmark is set to true", () => {
    expect(
      buildSearchParamsString(
        {
          param1: "value1",
          param2: 2,
          emptyStringWillBeSkipped: "",
          undefinedWillBeSkipped: undefined,
        },
        true
      )
    ).toEqual("param1=value1&param2=2");
  });

  test("should always return an empty string when searchParmsObj is empty", () => {
    expect(buildSearchParamsString({})).toEqual("");
    expect(buildSearchParamsString({ thisWillBeSkipped: undefined })).toEqual(
      ""
    );
    expect(buildSearchParamsString({}, true)).toEqual("");
  });
});
