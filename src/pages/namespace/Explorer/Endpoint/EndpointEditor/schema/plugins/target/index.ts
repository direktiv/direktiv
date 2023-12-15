import { InstantResposeFormSchema } from "./InstantResponse";
import { TargetFlowFormSchema } from "./TargetFlow";
import { z } from "zod";

export const targetPluginTypes = {
  instantResponse: "instant-response",
  targetFlow: "target-flow",
} as const;

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResposeFormSchema,
  TargetFlowFormSchema,
]);
