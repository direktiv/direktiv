import { describe, expect, test } from "vitest";
import { isValidDecimalLiteral, isValidDurationLiteral } from "../utils";

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

    test("decimal bound checks are intentionally loose", () => {
      expect(isValidDecimalLiteral("922337203685477.5808")).toBe(true);
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

    test("duration overflow checks are intentionally loose", () => {
      expect(isValidDurationLiteral("1d9223372036854775807ms")).toBe(true);
    });
  });
});
