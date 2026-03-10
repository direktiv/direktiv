import { z } from "zod";

// @shadow_mode / @reason("temporary block")
export const AnnotationsSchema = z.record(z.union([z.string(), z.null()]));
