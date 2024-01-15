import { describe, expect, test } from "vitest";

import { mimeTypeToEditorSyntax } from "../utils";

describe("mimeTypeToEditorSyntax", () => {
  test("it must detect html", () => {
    expect(mimeTypeToEditorSyntax("text/html")).toBe("html");
  });

  test("it must detect css", () => {
    expect(mimeTypeToEditorSyntax("text/css")).toBe("css");
  });

  test("it must detect json", () => {
    expect(mimeTypeToEditorSyntax("application/json")).toBe("json");
  });

  test("it must detect shell", () => {
    expect(mimeTypeToEditorSyntax("application/x-sh")).toBe("shell");
    expect(mimeTypeToEditorSyntax("application/x-csh")).toBe("shell");
  });

  test("it must detect plaintext", () => {
    expect(mimeTypeToEditorSyntax("text/")).toBe("plaintext");
    expect(mimeTypeToEditorSyntax("text/whatever")).toBe("plaintext");
  });

  test("it must detect javascript", () => {
    expect(mimeTypeToEditorSyntax("application/javascript")).toBe("javascript");
    expect(mimeTypeToEditorSyntax("text/javascript")).toBe("javascript");
  });

  test("it must return undefined for unsuported mime types", () => {
    expect(mimeTypeToEditorSyntax("unsupported")).toBe(undefined);
  });
});
