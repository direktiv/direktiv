import { describe, expect, test } from "vitest";

import { fileNameSchema } from "../schema/node";

describe("file name schema", () => {
  describe("valid", () => {
    test("all lowercase", () => {
      expect(fileNameSchema.safeParse("foldername").success).toBe(true);
    });

    test("may container a - in the middle", () => {
      expect(fileNameSchema.safeParse("folder-name").success).toBe(true);
    });

    test("may contain a dot in the middle", () => {
      expect(fileNameSchema.safeParse("folder.name").success).toBe(true);
    });

    test("may contain a _ in the middle", () => {
      expect(fileNameSchema.safeParse("folder_name").success).toBe(true);
    });

    test("may contain numbers if not the first character", () => {
      expect(fileNameSchema.safeParse("a123").success).toBe(true);
    });
  });

  describe("invalid", () => {
    test("must not contain only number", () => {
      expect(fileNameSchema.safeParse("123").success).toBe(false);
    });

    test("must not contain any slashes", () => {
      expect(fileNameSchema.safeParse("some/folder").success).toBe(false);
      expect(fileNameSchema.safeParse("some/").success).toBe(false);
      expect(fileNameSchema.safeParse("/folder").success).toBe(false);
    });

    test("must not contain uppercase characters middle", () => {
      expect(fileNameSchema.safeParse("fOldername").success).toBe(false);
    });

    test("must not end with various characters other than lowercase characters or a digit", () => {
      ["A", "-", ".", "_", "🙃"].forEach((char) => {
        expect(fileNameSchema.safeParse(`abc${char}`).success).toBe(false);
      });
    });

    test("must not start with various characters other than lowercase characters", () => {
      ["A", "1", ".", "_", "-", ".", "🙃"].forEach((char) => {
        expect(fileNameSchema.safeParse(`${char}abc`).success).toBe(false);
      });
    });
  });
});
