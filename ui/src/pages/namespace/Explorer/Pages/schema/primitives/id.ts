import { z } from "zod";

export const Id = z.string().min(1);
