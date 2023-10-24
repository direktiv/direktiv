import { SquareAsterisk } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

export const useNotificationConfig = () => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  if (!namespace) return null;
  return {
    secret: {
      icon: SquareAsterisk,
      description: (count: number) =>
        t("components.notificationMenu.hasIssues.secrets.description", {
          count,
        }),
      href: pages.settings.createHref({
        namespace,
      }),
    },
  } as const;
};
