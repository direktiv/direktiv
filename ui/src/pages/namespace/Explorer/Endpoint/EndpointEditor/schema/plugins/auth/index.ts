import { filterAvailablePlugins } from "../utils";

export const webhookAuthPluginTypes = {
  githubWebhookAuth: { name: "github-webhook-auth", enterpriseOnly: false },
  gitlabWebhookAuth: { name: "gitlab-webhook-auth", enterpriseOnly: false },
  slackWebhookAuth: { name: "slack-webhook-auth", enterpriseOnly: false },
} as const;

export const authPluginTypes = {
  basicAuth: { name: "basic-auth", enterpriseOnly: false },
  keyAuth: { name: "key-auth", enterpriseOnly: false },
  ...webhookAuthPluginTypes,
} as const;

export const availablePlugins = Object.values(authPluginTypes).filter(
  filterAvailablePlugins
);
