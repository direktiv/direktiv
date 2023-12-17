import { InstantResponseFormSchema } from "./InstantResponse";
import { TargetFlowFormSchema } from "./TargetFlow";
import { z } from "zod";

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResponseFormSchema,
  TargetFlowFormSchema,
]);
