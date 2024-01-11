export const authPluginTypes = {
  basicAuth: { name: "basic-auth", enterpriseOnly: false },
  githubWebhookAuth: { name: "github-webhook-auth", enterpriseOnly: false },
  keyAuth: { name: "key-auth", enterpriseOnly: false },
} as const;

const isEnterprise = !!process.env.VITE?.VITE_IS_ENTERPRISE;

export const availablePlugins = Object.values(authPluginTypes).filter(
  (plugin) => (isEnterprise ? true : plugin.enterpriseOnly === false)
);
