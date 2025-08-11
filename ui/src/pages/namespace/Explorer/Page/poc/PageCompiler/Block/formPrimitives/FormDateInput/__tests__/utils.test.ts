import { describe, expect, test } from "vitest";

import { parseStringToDate } from "../utils";

describe("parseStringToDate", () => {
  test("it should return undefined for empty string", () => {
    const result = parseStringToDate("");
    expect(result).toEqual(undefined);
  });

  test("it should return undefined for whitespace-only string", () => {
    const result = parseStringToDate("   ");
    expect(result).toEqual(undefined);
  });

  test("it should return undefined for invalid date string", () => {
    expect(parseStringToDate("invalid-date")).toEqual(undefined);
    expect(parseStringToDate("  ")).toEqual(undefined);
  });

  test("it should return a valid Date for ISO date string", () => {
    const dateString = "2023-12-25T10:30:00.000Z";
    const result = parseStringToDate(dateString);
    expect(result).toBeInstanceOf(Date);
    expect(result?.toISOString()).toEqual(dateString);
  });

  test("it should return a valid Date for simple date string", () => {
    const dateString = "2023-12-25";
    const result = parseStringToDate(dateString);
    expect(result).toBeInstanceOf(Date);
    expect(result?.getFullYear()).toEqual(2023);
    expect(result?.getMonth()).toEqual(11);
    expect(result?.getDate()).toEqual(25);
  });

  test("it should return a valid Date for US date format", () => {
    const dateString = "12/25/2023";
    const result = parseStringToDate(dateString);
    expect(result).toBeInstanceOf(Date);
    expect(result?.getFullYear()).toEqual(2023);
    expect(result?.getMonth()).toEqual(11);
    expect(result?.getDate()).toEqual(25);
  });
});
