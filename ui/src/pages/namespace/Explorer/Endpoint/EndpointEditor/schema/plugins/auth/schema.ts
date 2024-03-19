import { BasicAuthFormSchema } from "./basicAuth";
import { KeyAuthFormSchema } from "./keyAuth";
import { WebhookAuthFormSchema } from "./webhookAuth";
import { z } from "zod";

export const AuthPluginFormSchema = z.discriminatedUnion("type", [
  BasicAuthFormSchema,
  WebhookAuthFormSchema,
  KeyAuthFormSchema,
]);

export type AuthPluginFormSchemaType = z.infer<typeof AuthPluginFormSchema>;
