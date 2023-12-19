import { authPluginTypes } from ".";
import { z } from "zod";

export const GithubWebhookAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.githubWebhookAuth),
  configuration: z.object({}),
});

export type GithubWebhookAuthFormSchemaType = z.infer<
  typeof GithubWebhookAuthFormSchema
>;
