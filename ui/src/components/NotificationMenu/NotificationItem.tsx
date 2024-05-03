import { NotificationClose, NotificationMessage } from "~/design/Notification/";

import { FC } from "react";
import { Link } from "react-router-dom";
import { NotificationSchemaType } from "~/api/notifications/schema";
import { useNotificationConfig } from "./config";

export const NotificationItem: FC<NotificationSchemaType> = ({
  type,
  count,
}) => {
  const config = useNotificationConfig({ type, count });

  if (!config) return null;

  const { description, href, icon: Icon } = config;

  return (
    <NotificationClose
      className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3"
      asChild
    >
      <Link to={href}>
        <NotificationMessage text={description} icon={Icon} />
      </Link>
    </NotificationClose>
  );
};
