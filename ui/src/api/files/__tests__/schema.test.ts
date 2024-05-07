import { describe, expect, test } from "vitest";

import { FileNameSchema } from "../schema";

describe("file name schema", () => {
  describe("valid", () => {
    test("all lowercase", () => {
      expect(FileNameSchema.safeParse("foldername").success).toBe(true);
    });

    test("may container a - in the middle", () => {
      expect(FileNameSchema.safeParse("folder-name").success).toBe(true);
    });

    test("may contain a dot in the middle", () => {
      expect(FileNameSchema.safeParse("folder.name").success).toBe(true);
    });

    test("may contain a _ in the middle", () => {
      expect(FileNameSchema.safeParse("folder_name").success).toBe(true);
    });

    test("may contain numbers if not the first character", () => {
      expect(FileNameSchema.safeParse("a123").success).toBe(true);
    });
  });

  describe("invalid", () => {
    test("must not contain only number", () => {
      expect(FileNameSchema.safeParse("123").success).toBe(false);
    });

    test("must not contain any slashes", () => {
      expect(FileNameSchema.safeParse("some/folder").success).toBe(false);
      expect(FileNameSchema.safeParse("some/").success).toBe(false);
      expect(FileNameSchema.safeParse("/folder").success).toBe(false);
    });

    test("must not contain uppercase characters middle", () => {
      expect(FileNameSchema.safeParse("fOldername").success).toBe(false);
    });

    test("must not end with various characters other than lowercase characters or a digit", () => {
      ["A", "-", ".", "_", "🙃"].forEach((char) => {
        expect(FileNameSchema.safeParse(`abc${char}`).success).toBe(false);
      });
    });

    test("must not start with various characters other than lowercase characters", () => {
      ["A", "1", ".", "_", "-", ".", "🙃"].forEach((char) => {
        expect(FileNameSchema.safeParse(`${char}abc`).success).toBe(false);
      });
    });
  });
});
