import { describe, expect, test } from "vitest";

import { serviceNameSchema } from "../schema/services";

describe("service name schema", () => {
  describe("valid", () => {
    test("all lowercase", () => {
      expect(serviceNameSchema.safeParse("servicename").success).toBe(true);
    });

    test("may contain a - in the middle", () => {
      expect(serviceNameSchema.safeParse("service-name").success).toBe(true);
    });

    test("may contain numbers if not the first character", () => {
      expect(serviceNameSchema.safeParse("a123").success).toBe(true);
    });
  });

  describe("invalid", () => {
    test("must not contain only numbers", () => {
      expect(serviceNameSchema.safeParse("123").success).toBe(false);
    });

    test("must not contain any dots", () => {
      expect(serviceNameSchema.safeParse("service.name").success).toBe(false);
      expect(serviceNameSchema.safeParse("servicename.").success).toBe(false);
      expect(serviceNameSchema.safeParse(".servicename").success).toBe(false);
    });

    test("must not contain a _ ", () => {
      expect(serviceNameSchema.safeParse("service_name").success).toBe(false);
      expect(serviceNameSchema.safeParse("servicename_").success).toBe(false);
      expect(serviceNameSchema.safeParse("_servicename").success).toBe(false);
    });

    test("must not contain any slashes", () => {
      expect(serviceNameSchema.safeParse("some/folder").success).toBe(false);
      expect(serviceNameSchema.safeParse("some/").success).toBe(false);
      expect(serviceNameSchema.safeParse("/folder").success).toBe(false);
    });

    test("must not contain uppercase characters middle", () => {
      expect(serviceNameSchema.safeParse("fOldername").success).toBe(false);
    });

    test("must not end with characters other than lowercase letters", () => {
      ["A", "-", ".", "_", "🙃, 1"].forEach((char) => {
        expect(serviceNameSchema.safeParse(`abc${char}`).success).toBe(false);
      });
    });

    test("must not start with characters other than lowercase letters", () => {
      ["A", "1", ".", "_", "-", ".", "🙃"].forEach((char) => {
        expect(serviceNameSchema.safeParse(`${char}abc`).success).toBe(false);
      });
    });
  });
});
