import { LucideIcon, SquareAsterisk } from "lucide-react";

import { LinkComponentProps } from "@tanstack/react-router";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

export type NotificationConfig = {
  linkProps: LinkComponentProps;
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
        linkProps: { to: "/n/$namespace/settings", params: { namespace } },
      };
    default:
      return null;
  }
};
