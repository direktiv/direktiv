import { describe, expect, test } from "vitest";

import { parseDataUrl } from "../utils";

describe("parseDataUrl", () => {
  test("should return a base64 string and mime type from a valid data url", () => {
    const validDataUrl =
      "data:application/json;base64,ewogICAgImNvb2xWaWRlbyI6ICJodHRwczovL3lvdXR1LmJlL29IZzVTSllSSEEwP3NpPXlRLVFLMEE1RlBiSG5rZDgiCn0=";

    expect(parseDataUrl(validDataUrl)).toEqual({
      base64String:
        "ewogICAgImNvb2xWaWRlbyI6ICJodHRwczovL3lvdXR1LmJlL29IZzVTSllSSEEwP3NpPXlRLVFLMEE1RlBiSG5rZDgiCn0=",
      mimeType: "application/json",
    });
  });

  describe("should return null when the data url is not a valid base64 string", () => {
    test("empty string", () => {
      expect(parseDataUrl("")).toEqual(null);
    });

    test("empty base64 string", () => {
      expect(parseDataUrl("data:application/json;base64,")).toEqual(null);
    });

    test("empty mime type", () => {
      expect(
        parseDataUrl(
          "data:;base64,ewogICAgImNvb2xWaWRlbyI6ICJodHRwczovL3lvdXR1LmJlL29IZzVTSllSSEEwP3NpPXlRLVFLMEE1RlBiSG5rZDgiCn0="
        )
      ).toEqual(null);
    });
  });
});
