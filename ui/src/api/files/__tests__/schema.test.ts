import { describe, expect, test } from "vitest";
import { getFilenameFromPath, getParentFromPath } from "../schema";

describe("getFilenameFromPath()", () => {
  test("file at root level", () => {
    expect(getFilenameFromPath("/action.yaml")).toBe("action.yaml");
  });
  test("file in directory", () => {
    expect(getFilenameFromPath("/foobar/nested.yaml")).toBe("nested.yaml");
  });
  test("directory", () => {
    expect(getFilenameFromPath("/examples")).toBe("examples");
  });
  test("nested directory", () => {
    expect(getFilenameFromPath("/examples/nested")).toBe("nested");
  });
});

describe("getParentFromPath()", () => {
  test("file at root level", () => {
    expect(getParentFromPath("/action.yaml")).toBe("/");
  });
  test("file in directory", () => {
    expect(getParentFromPath("/foobar/nested.yaml")).toBe("/foobar");
  });
  test("directory at root level", () => {
    expect(getParentFromPath("/examples")).toBe("/");
  });
  test("nested directory", () => {
    expect(getParentFromPath("/examples/nested")).toBe("/examples");
  });
});
