import { InstantResponseFormSchema } from "./InstantResponse";
import { TargetFlowFormSchema } from "./TargetFlow";
import { TargetFlowVarFormSchema } from "./TargetFlowVar";
import { TargetNamespaceFileFormSchema } from "./TargetNamespaceFile";
import { TargetNamespaceVarFormSchema } from "./TargetNamespaceVar";
import { z } from "zod";

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResponseFormSchema,
  TargetFlowFormSchema,
  TargetFlowVarFormSchema,
  TargetNamespaceFileFormSchema,
  TargetNamespaceVarFormSchema,
]);
