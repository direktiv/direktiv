export const authPluginTypes = {
  basicAuth: "basic-auth",
  githubWebhookAuth: "github-webhook-auth",
  keyAuth: "key-auth",
} as const;

export const availablePlugins = Object.values(authPluginTypes);
