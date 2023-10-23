import { FC, PropsWithChildren } from "react";
import { Loader2, LucideIcon } from "lucide-react";

import { DropdownMenuSeparator } from "../Dropdown";

export const NotificationTitle: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationText: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-11 dark:text-gray-dark-11">
    {children}
  </div>
);

export const NotificationLoading: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex items-center py-1.5 px-2">
    <Loader2 className="h-4 w-4 animate-spin" />
    <div className="px-2 py-1.5 text-sm font-medium text-gray-11 dark:text-gray-dark-11">
      {children}
    </div>
  </div>
);

export function NotificationMessage({
  title,
  text,
  icon: Icon,
}: {
  title: string;
  text: string;
  icon: LucideIcon;
}) {
  return (
    <div className="flex flex-col">
      <div>
        <NotificationTitle>{title}</NotificationTitle>
        <DropdownMenuSeparator className="-mx-1 my-1 h-px bg-gray-3 dark:bg-gray-dark-3"></DropdownMenuSeparator>
      </div>
      <div className="flex items-center py-1.5 px-2">
        <Icon
          className="h-4 w-4 text-gray-11 dark:text-gray-dark-11"
          aria-hidden="true"
        />
        <NotificationText>{text}</NotificationText>
      </div>
    </div>
  );
}
