import { FC } from "react";
import { LucideIcon } from "lucide-react";

export type NotificationItemType = {
  icon: LucideIcon;
  description: string;
  href: string;
};

export const NotificationItem: FC<NotificationItemType> = ({
  href,
  description,
  icon: Icon,
}) => (
  <div>
    <div>{href}</div>
    <div>{description}</div>
    <Icon />
  </div>
);
