import { Mutation } from "../dataFetching/mutation";
import { z } from "zod";

export const Button = z.object({
  type: z.literal("button"),
  data: z.object({
    label: z.string().min(1),
    submit: Mutation,
  }),
});

export type ButtonType = z.infer<typeof Button>;
