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
import { useGroupNotifications } from "./config";
import { useNotifications } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useTranslation } from "react-i18next";

const NotificationMenu = ({ className }: { className?: string }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNotifications();

  const showIndicator = !!data?.data.length;
  const notificationItems = useGroupNotifications(data);

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
          notificationItems.map((item, index) => {
            const isLastListItem = index === notificationItems.length - 1;
            return (
              <Fragment key={index}>
                <NotificationItem {...item} />
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
