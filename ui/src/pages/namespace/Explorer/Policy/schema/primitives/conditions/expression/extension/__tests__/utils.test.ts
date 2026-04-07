import {
  cedarIpLiteralSchema,
  isValidDecimalLiteral,
  isValidDurationLiteral,
} from "../utils";
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
      expect(isValidDecimalLiteral("922337203685477.5808")).toBe(false);
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

  describe("cedarIpLiteralSchema", () => {
    test("accepts valid Cedar IP literals", () => {
      expect(cedarIpLiteralSchema.safeParse("127.0.0.1").success).toBe(true);
      expect(cedarIpLiteralSchema.safeParse("127.0.0.1/24").success).toBe(true);
      expect(cedarIpLiteralSchema.safeParse("::1").success).toBe(true);
      expect(cedarIpLiteralSchema.safeParse("ffee::/64").success).toBe(true);
    });

    test("rejects invalid Cedar IP literals", () => {
      expect(cedarIpLiteralSchema.safeParse("380.0.0.1").success).toBe(false);
      expect(cedarIpLiteralSchema.safeParse("127.0.0.1/8/24").success).toBe(
        false
      );
      expect(cedarIpLiteralSchema.safeParse("fee::/64::1").success).toBe(false);
    });
  });
});
