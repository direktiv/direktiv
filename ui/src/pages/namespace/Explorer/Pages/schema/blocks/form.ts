import { Blocks, BlocksType } from ".";

import { Mutation, MutationType } from "../misc/mutation";
import { z } from "zod";

export type FormType = {
  type: "form";
  data: {
    trigger: BlocksType["trigger"];
    apiRequest: MutationType;
    blocks: BlocksType["all"][];
  };
};

export const Form = z.object({
  type: z.literal("form"),
  data: z.object({
    trigger: Blocks.trigger,
    apiRequest: Mutation,
    blocks: z.array(Blocks.all),
  }),
}) satisfies z.ZodType<FormType>;
