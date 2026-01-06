import { InstantResponseFormSchema } from "./instantResponse";
import { TargetEventFormSchema } from "./targetEvent";
import { TargetFlowFormSchema } from "./targetFlow";
import { TargetFlowVarFormSchema } from "./targetFlowVar";
import { TargetNamespaceFileFormSchema } from "./targetNamespaceFile";
import { TargetNamespaceVarFormSchema } from "./targetNamespaceVar";
import { TargetPageFormSchema } from "./targetPage";
import { z } from "zod";

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResponseFormSchema,
  TargetFlowFormSchema,
  TargetPageFormSchema,
  TargetFlowVarFormSchema,
  TargetNamespaceFileFormSchema,
  TargetNamespaceVarFormSchema,
  TargetEventFormSchema,
]);

export type TargetPluginFormSchemaType = z.infer<typeof TargetPluginFormSchema>;
