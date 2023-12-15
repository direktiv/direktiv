import { InstantResposeFormSchema } from "./InstantResponse";
import { TargetFlowFormSchema } from "./TargetFlow";
import { z } from "zod";

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResposeFormSchema,
  TargetFlowFormSchema,
]);
