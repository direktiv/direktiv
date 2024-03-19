import { authPluginTypes } from ".";
import { z } from "zod";

export const webhookAuthPluginNames = [
  authPluginTypes.githubWebhookAuth.name,
  authPluginTypes.gitlabWebhookAuth.name,
  authPluginTypes.slackWebhookAuth.name,
] as const;

export const WebhookAuthFormSchema = z.object({
  type: z.enum(webhookAuthPluginNames),
  configuration: z.object({
    secret: z.string().nonempty(),
  }),
});

export type WebhookAuthFormSchemaType = z.infer<typeof WebhookAuthFormSchema>;
