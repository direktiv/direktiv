import { describe, expect, test } from "vitest";

import { DirektivPagesSchema } from "..";
import simple from "./examples/simple";

describe("Direktiv Pages zod schema", () => {
  test("The simple direktiv pages file is valid", () => {
    expect(DirektivPagesSchema.safeParse(simple).success).toBe(true);
    expect(DirektivPagesSchema.parse(simple)).toEqual(simple);
  });
});
