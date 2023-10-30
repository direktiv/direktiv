import { LintSchemaType } from "~/api/namespaceLinting/schema";
import { NotificationItemType } from "./NotificationItem";
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

export const useGroupNotifications = (data: LintSchemaType | undefined) => {
  const notificationConfig = useNotificationConfig();
  const notificationTypes = Object.entries(notificationConfig ?? {});
  return notificationTypes
    .map(([notificationType, notificationConfig]) => {
      const matchingNotifications = data?.issues.filter(
        (issue) => notificationType === issue.type
      );

      if (!matchingNotifications || matchingNotifications.length === 0) {
        return null;
      }

      const { icon, description, href } = notificationConfig;
      return {
        icon,
        description: description(matchingNotifications.length),
        href,
      };
    })
    .filter((item) => item !== null) as NotificationItemType[];
};
