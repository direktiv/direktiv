import { isValidDecimalLiteral, isValidDurationLiteral } from "../utils";
import { describe, expect, test } from "vitest";

describe("extension utils", () => {
  describe("isValidDecimalLiteral", () => {
    test("accepts valid Cedar decimals", () => {
      expect(isValidDecimalLiteral("1.0")).toBe(true);
      expect(isValidDecimalLiteral("0.1234")).toBe(true);
      expect(isValidDecimalLiteral("-922337203685477.5808")).toBe(true);
    });

    test("rejects invalid Cedar decimals", () => {
      expect(isValidDecimalLiteral("1234")).toBe(false);
      expect(isValidDecimalLiteral("1.")).toBe(false);
      expect(isValidDecimalLiteral("0.12345")).toBe(false);
    });
  });

  describe("isValidDurationLiteral", () => {
    test("accepts valid Cedar durations", () => {
      expect(isValidDurationLiteral("2h30m")).toBe(true);
      expect(isValidDurationLiteral("-1d12h")).toBe(true);
      expect(isValidDurationLiteral("1h30m45s")).toBe(true);
    });

    test("rejects invalid Cedar durations", () => {
      expect(isValidDurationLiteral("1s1d")).toBe(false);
      expect(isValidDurationLiteral("1s1s")).toBe(false);
      expect(isValidDurationLiteral("1d2h3m4s5ms6")).toBe(false);
    });
  });
});
