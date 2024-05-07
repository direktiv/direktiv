import { LucideIcon, SquareAsterisk } from "lucide-react";

import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

export type NotificationConfig = {
  href: string;
  description: string;
  icon: LucideIcon;
};

export const useNotificationConfig = ({
  type,
  count,
}: {
  type: string;
  count: number;
}): NotificationConfig | null => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  if (!namespace) return null;

  switch (type) {
    case "uninitialized_secrets":
      return {
        icon: SquareAsterisk,
        description: t(
          "components.notificationMenu.hasIssues.secrets.description",
          {
            count,
          }
        ),
        href: pages.settings.createHref({
          namespace,
        }),
      };
    default:
      return null;
  }
};
