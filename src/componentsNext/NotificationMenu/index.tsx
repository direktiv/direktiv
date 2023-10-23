import {
  NotificationLoading,
  NotificationMessage,
  NotificationText,
} from "~/design/Notification/NotificationModal";

import { Link } from "react-router-dom";
import Notification from "~/design/Notification";
import { Settings } from "lucide-react";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useTranslation } from "react-i18next";

interface NotificationMenuProps {
  className?: string;
}

const useNotificationConfig = () => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  if (!namespace) return null;
  return {
    secret: {
      icon: Settings,
      title: t("components.notificationMenu.hasIssues.secrets.title"),
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

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNamespaceLinting();
  const notificationConfig = useNotificationConfig();
  const showIndicator = !!data?.issues.length;

  const textLoading = t("components.notificationMenu.isLoading.text");
  const textNoIssues = t("components.notificationMenu.noIssues.text");

  const namespace = useNamespace();
  if (!namespace) return null;

  const possibleNotifications = Object.entries(notificationConfig ?? {});

  return (
    <div className={twMergeClsx("self-end text-right", className)}>
      <Notification showIndicator={showIndicator}>
        {isLoading && <NotificationLoading>{textLoading}</NotificationLoading>}
        {showIndicator && !isLoading && (
          <div>
            {possibleNotifications.map(
              ([notificationType, notificationConfig]) => {
                const matchingNotification = data.issues.filter(
                  (issue) => notificationType === issue.type
                );

                if (matchingNotification.length <= 0) {
                  return null;
                }

                return (
                  <Link to={notificationConfig.href} key={notificationType}>
                    <NotificationMessage
                      title={notificationConfig.title}
                      text={notificationConfig.description(
                        matchingNotification.length
                      )}
                      icon={notificationConfig.icon}
                    />
                  </Link>
                );
              }
            )}
          </div>
        )}
        {!showIndicator && !isLoading && (
          <div className="flex items-center py-1.5 px-2">
            <NotificationText>{textNoIssues}</NotificationText>
          </div>
        )}
      </Notification>
    </div>
  );
};

export default NotificationMenu;
