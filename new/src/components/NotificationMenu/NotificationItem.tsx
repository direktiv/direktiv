import { NotificationClose, NotificationMessage } from "~/design/Notification/";

import { FC } from "react";
import { Link } from "react-router-dom";
import { LucideIcon } from "lucide-react";

export type NotificationItemType = {
  href: string;
  description: string;
  icon: LucideIcon;
};

export const NotificationItem: FC<NotificationItemType> = ({
  href,
  description,
  icon: Icon,
}) => (
  <NotificationClose
    className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3"
    asChild
  >
    <Link to={href}>
      <NotificationMessage text={description} icon={Icon} />
    </Link>
  </NotificationClose>
);
