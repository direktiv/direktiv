import { NotificationItemType } from "./NotificationItem";
import { NotificationListSchemaType } from "~/api/notifications/schema";
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

export const useGroupNotifications = (
  data: NotificationListSchemaType | undefined
) => {
  const notificationConfig = useNotificationConfig();
  const notificationTypes = Object.entries(notificationConfig ?? {});
  return notificationTypes
    .map(([notificationType, notificationConfig]) => {
      const matchingNotifications = data?.data.filter(
        (issue) => notificationType === issue.type
      );

      if (!matchingNotifications || matchingNotifications.length === 0) {
        return null;
      }

      const { href, description, icon } = notificationConfig;
      return {
        href,
        description: description(matchingNotifications.length),
        icon,
      };
    })
    .filter((item) => item !== null) as NotificationItemType[];
};
