import { describe, expect, test } from "vitest";

import { DirektivPagesSchema } from "..";
import complex from "./examples/complex";
import simple from "./examples/simple";

describe("Direktiv Pages zod schema", () => {
  test("The simple direktiv pages example file is valid", () => {
    expect(DirektivPagesSchema.safeParse(simple).success).toBe(true);
    expect(DirektivPagesSchema.parse(simple)).toEqual(simple);
  });

  test("The complex direktiv pages exampe file is valid", () => {
    expect(DirektivPagesSchema.safeParse(complex).success).toBe(true);
    expect(DirektivPagesSchema.parse(complex)).toEqual(complex);
  });
});
