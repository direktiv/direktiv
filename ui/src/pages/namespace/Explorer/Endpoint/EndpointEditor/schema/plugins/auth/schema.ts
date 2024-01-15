import { BasicAuthFormSchema } from "./basicAuth";
import { GithubWebhookAuthFormSchema } from "./githubWebhookAuth";
import { KeyAuthFormSchema } from "./keyAuth";
import { z } from "zod";

export const AuthPluginFormSchema = z.discriminatedUnion("type", [
  BasicAuthFormSchema,
  GithubWebhookAuthFormSchema,
  KeyAuthFormSchema,
]);

export type AuthPluginFormSchemaType = z.infer<typeof AuthPluginFormSchema>;
