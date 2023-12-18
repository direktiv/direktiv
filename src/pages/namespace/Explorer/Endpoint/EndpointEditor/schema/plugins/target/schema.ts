import { InstantResponseFormSchema } from "./instantResponse";
import { TargetFlowFormSchema } from "./targetFlow";
import { TargetFlowVarFormSchema } from "./targetFlowVar";
import { TargetNamespaceFileFormSchema } from "./targetNamespaceFile";
import { TargetNamespaceVarFormSchema } from "./targetNamespaceVar";
import { z } from "zod";

export const TargetPluginFormSchema = z.discriminatedUnion("type", [
  InstantResponseFormSchema,
  TargetFlowFormSchema,
  TargetFlowVarFormSchema,
  TargetNamespaceFileFormSchema,
  TargetNamespaceVarFormSchema,
]);
