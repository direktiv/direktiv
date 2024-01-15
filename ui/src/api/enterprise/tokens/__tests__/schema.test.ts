import { describe, expect, test } from "vitest";

import { ISO8601durationSchema } from "../schema";

describe("ISO 8601 duration schema", () => {
  describe("valid", () => {
    test("1 year", () => {
      expect(ISO8601durationSchema.safeParse("P1Y").success).toBe(true);
    });

    test("2 months", () => {
      expect(ISO8601durationSchema.safeParse("P2M").success).toBe(true);
    });

    test("3 days", () => {
      expect(ISO8601durationSchema.safeParse("P3D").success).toBe(true);
    });

    test("4 hours", () => {
      expect(ISO8601durationSchema.safeParse("PT4H").success).toBe(true);
    });

    test("30 minutes", () => {
      expect(ISO8601durationSchema.safeParse("PT30M").success).toBe(true);
    });

    test("45 seconds", () => {
      expect(ISO8601durationSchema.safeParse("PT45S").success).toBe(true);
    });

    test("1 year, 2 months, 3 days, 4 hours, 5 minutes, 6 seconds", () => {
      expect(ISO8601durationSchema.safeParse("P1Y2M3DT4H5M6S").success).toBe(
        true
      );
    });

    test("7 weeks", () => {
      expect(ISO8601durationSchema.safeParse("P7W").success).toBe(true);
    });

    test("10 years, 6 months", () => {
      expect(ISO8601durationSchema.safeParse("P10Y6M").success).toBe(true);
    });

    test("1 year, 6 months, 7 days, 8 hours", () => {
      expect(ISO8601durationSchema.safeParse("P1Y6M7DT8H").success).toBe(true);
    });

    test("1 day, 12 hours, 30 minutes", () => {
      expect(ISO8601durationSchema.safeParse("P1DT12H30M").success).toBe(true);
    });

    test("4 years, 3 months, 2 days", () => {
      expect(ISO8601durationSchema.safeParse("P4Y3M2D").success).toBe(true);
    });

    test("5 months, 15 days, 6 hours", () => {
      expect(ISO8601durationSchema.safeParse("P5M15DT6H").success).toBe(true);
    });

    test("20 minutes, 45 seconds", () => {
      expect(ISO8601durationSchema.safeParse("PT20M45S").success).toBe(true);
    });

    test("2 years, 3 days", () => {
      expect(ISO8601durationSchema.safeParse("P2Y3D").success).toBe(true);
    });

    test("8 weeks, 3 days, 2 hours", () => {
      expect(ISO8601durationSchema.safeParse("P8W3DT2H").success).toBe(true);
    });

    test("1 hour, 15 minutes, 30 seconds", () => {
      expect(ISO8601durationSchema.safeParse("PT1H15M30S").success).toBe(true);
    });

    test("9 years, 6 months, 7 days, 4 hours", () => {
      expect(ISO8601durationSchema.safeParse("P9Y6M7DT4H").success).toBe(true);
    });

    test("45 minutes, 20 seconds", () => {
      expect(ISO8601durationSchema.safeParse("PT45M20S").success).toBe(true);
    });

    test("11 years, 10 months, 9 days, 8 hours, 7 minutes", () => {
      expect(ISO8601durationSchema.safeParse("P11Y10M9DT8H7M").success).toBe(
        true
      );
    });
  });

  describe("invalid", () => {
    test("no metric given", () => {
      expect(ISO8601durationSchema.safeParse("PT45").success).toBe(false);
      expect(ISO8601durationSchema.safeParse("T45").success).toBe(false);
    });
  });

  test("invalid characters", () => {
    expect(
      ISO8601durationSchema.safeParse("P1Y2M3DT4H5M6S!").success,
      "! at the end"
    ).toBe(false);

    expect(
      ISO8601durationSchema.safeParse("P1Y2M3D@T4H5M6S").success,
      "@ in the middle"
    ).toBe(false);

    test("missing 'T' separator for time", () => {
      expect(ISO8601durationSchema.safeParse("P1Y2M3D4H5M6S").success).toBe(
        false
      );
    });

    test("missing 'P' at the beginning", () => {
      expect(ISO8601durationSchema.safeParse("1Y2M3DT4H5M6S").success).toBe(
        false
      );
    });

    test("invalid order of components", () => {
      expect(
        ISO8601durationSchema.safeParse("PT4H3D").success,
        "Days should come before hours"
      ).toBe(false);
      expect(
        ISO8601durationSchema.safeParse("P1H2D").success,
        "Hours cannot come before days"
      ).toBe(false);
    });

    test("empty string", () => {
      expect(ISO8601durationSchema.safeParse("").success).toBe(false);
    });

    test("only P and T", () => {
      expect(ISO8601durationSchema.safeParse("P").success).toBe(false);
      expect(ISO8601durationSchema.safeParse("T").success).toBe(false);
    });

    test("multiple P separators", () => {
      expect(ISO8601durationSchema.safeParse("P1PY1M").success).toBe(false);
    });

    test("multiple T separators", () => {
      expect(ISO8601durationSchema.safeParse("PT1HT1M").success).toBe(false);
    });

    test("invalid time components", () => {
      expect(
        ISO8601durationSchema.safeParse("PT60M").success,
        "60 minutes is not valid; it should be PT1H"
      ).toBe(false);

      expect(
        ISO8601durationSchema.safeParse("PT24H").success,
        "24 hours is not valid; it should be P1D."
      ).toBe(false);
    });
  });
});
