import { Check, LucideIcon, SquareAsterisk } from "lucide-react";
import {
  Notification,
  NotificationClose,
  NotificationLoading,
  NotificationMenuSeparator,
  NotificationMessage,
  NotificationTitle,
} from "~/design/Notification/";

import { Link } from "react-router-dom";
import { twMergeClsx } from "~/util/helpers";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNotificationConfig } from "./config";
import { useTranslation } from "react-i18next";

function NotificationItem({
  href,
  description,
  icon: Icon,
}: {
  href: string;
  description: string;
  icon: LucideIcon;
}) {
  return (
    <div>
      <div>{href}</div>
      <div>{description}</div>
      <Icon />
    </div>
  );
}

interface NotificationMenuProps {
  className?: string;
}

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNamespaceLinting();
  const notificationConfig = useNotificationConfig();
  const showIndicator = !!data?.issues.length;

  const notificationTypes = Object.entries(notificationConfig ?? {});

  const notificationItems = notificationTypes.map(
    ([notificationType, notificationConfig], index, srcArr) => {
      const matchingNotifications = data?.issues.filter(
        (issue) => notificationType === issue.type
      );
      if (
        matchingNotifications == undefined ||
        matchingNotifications.length <= 0
      ) {
        return null;
      }

      let { icon, description, count, href } = srcArr[index];
    }
  );

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
        {showIndicator && notificationItems != null ? (
          notificationItems.map((item, index) => {
            const isLastListItem = index === notificationItems.length - 1;
            return (
              <div key="">
                <NotificationItem
                  icon={item.icon}
                  href={item.href}
                  description={item.description}
                />
                {!isLastListItem && <NotificationMenuSeparator />}
              </div>
            );
          })
        ) : (
          <NotificationMessage
            text={t("components.notificationMenu.noIssues.description")}
            icon={Check}
          />
        )}
      </Notification>
    </div>
  );
};

export default NotificationMenu;
