import { isEnterprise } from "~/config/env/utils";
import { isPluginAvailable } from "../utils";

export const authPluginTypes = {
  basicAuth: { name: "basic-auth", enterpriseOnly: false },
  keyAuth: { name: "key-auth", enterpriseOnly: false },
  githubWebhookAuth: { name: "github-webhook-auth", enterpriseOnly: false },
  gitlabWebhookAuth: { name: "gitlab-webhook-auth", enterpriseOnly: false },
  slackWebhookAuth: { name: "slack-webhook-auth", enterpriseOnly: false },
} as const;

export const useAvailablePlugins = () =>
  Object.values(authPluginTypes).filter((plugin) =>
    isPluginAvailable(plugin, isEnterprise())
  );
