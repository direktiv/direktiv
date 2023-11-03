import {
  Notification,
  NotificationClose,
  NotificationLoading,
  NotificationMenuSeparator,
  NotificationMessage,
  NotificationTitle,
} from "~/design/Notification/";

import { Check } from "lucide-react";
import { Link } from "react-router-dom";
import { twMergeClsx } from "~/util/helpers";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNotificationConfig } from "./config";
import { useTranslation } from "react-i18next";

interface NotificationMenuProps {
  className?: string;
}

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNamespaceLinting();
  const notificationConfig = useNotificationConfig();
  const showIndicator = !!data?.issues.length;

  const possibleNotifications = Object.entries(notificationConfig ?? {});

  return (
    <div className={twMergeClsx("self-end text-right", className)}>
      <Notification showIndicator={showIndicator}>
        <NotificationTitle>
          {t("components.notificationMenu.title")}
        </NotificationTitle>
        <NotificationMenuSeparator />
        {isLoading && (
          <NotificationLoading>
            {t("components.notificationMenu.isLoading.description")}
          </NotificationLoading>
        )}
        {showIndicator ? (
          possibleNotifications.map(
            ([notificationType, notificationConfig], index, srcArr) => {
              const isLastListItem = index === srcArr.length - 1;
              const matchingNotification = data.issues.filter(
                (issue) => notificationType === issue.type
              );
              if (matchingNotification.length <= 0) {
                return null;
              }
              return (
                <div key={notificationType}>
                  <NotificationClose
                    className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3"
                    asChild
                  >
                    <Link to={notificationConfig.href}>
                      <NotificationMessage
                        text={notificationConfig.description(
                          matchingNotification.length
                        )}
                        icon={notificationConfig.icon}
                      />
                    </Link>
                  </NotificationClose>
                  {!isLastListItem && <NotificationMenuSeparator />}
                </div>
              );
            }
          )
        ) : (
          <NotificationMessage
            data-testid="notification-text"
            text={t("components.notificationMenu.noIssues.description")}
            icon={Check}
          />
        )}
      </Notification>
    </div>
  );
};

export default NotificationMenu;
