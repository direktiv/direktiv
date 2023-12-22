import { authPluginTypes } from ".";
import { z } from "zod";

export const GithubWebhookAuthFormSchema = z.object({
  type: z.literal(authPluginTypes.githubWebhookAuth),
  configuration: z.object({
    secret: z.string().nonempty(),
  }),
});

export type GithubWebhookAuthFormSchemaType = z.infer<
  typeof GithubWebhookAuthFormSchema
>;
