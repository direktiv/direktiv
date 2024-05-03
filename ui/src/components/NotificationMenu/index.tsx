import {
  Notification,
  NotificationLoading,
  NotificationMenuSeparator,
  NotificationMessage,
  NotificationTitle,
} from "~/design/Notification/";

import { Check } from "lucide-react";
import { Fragment } from "react";
import { NotificationItem } from "./NotificationItem";
import { twMergeClsx } from "~/util/helpers";
import { useNotifications } from "~/api/notifications/query/get";
import { useTranslation } from "react-i18next";

const NotificationMenu = ({ className }: { className?: string }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNotifications();

  const notifications = data?.data;
  const showIndicator = !!notifications?.length;

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
          notifications.map((notification, index) => {
            const isLastListItem = index === notifications.length - 1;
            return (
              <Fragment key={notification.type}>
                <NotificationItem {...notification} />
                {!isLastListItem && <NotificationMenuSeparator />}
              </Fragment>
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
