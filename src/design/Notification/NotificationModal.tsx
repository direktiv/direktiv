import { FC, PropsWithChildren } from "react";

import { Loader2 } from "lucide-react";
import { useTranslation } from "react-i18next";

export const NotificationNoresults: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationHasresultsTitle: FC<PropsWithChildren> = ({
  children,
}) => (
  <div className="px-2 py-1.5 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationHasresultsText: FC<PropsWithChildren> = ({
  children,
}) => (
  <div className="px-2 py-1.5 text-sm font-medium text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

export const NotificationHasresultsButton: FC<PropsWithChildren> = ({
  children,
}) => (
  <div className="relative flex cursor-pointer select-none items-end rounded-sm py-1.5 px-2 text-sm font-medium outline-none focus:bg-gray-3 data-[disabled]:pointer-events-none data-[disabled]:opacity-50 dark:focus:bg-gray-dark-3">
    {children}
  </div>
);

export const NotificationLoading: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex">
    <Loader2 className="h-5 animate-spin" />
    {children}
  </div>
);

const NotificationModal = ({
  showIndicator,
  isLoading,
}: {
  className?: string;
  showIndicator?: boolean;
  isLoading?: boolean;
}) => {
  const { t } = useTranslation();

  const textLoading = t("components.notificationMenu.isLoading.text");
  const textHasIssues = t("components.notificationMenu.hasIssues.secrets.text");
  const textNoIssues = t("components.notificationMenu.noIssues.text");

  if (isLoading) {
    return (
      <div className="">
        <Loader2 className="h-5 animate-spin" />
        {textLoading}
      </div>
    );
  }

  if (showIndicator) {
    return <div className="">{textHasIssues}</div>;
  }

  return <div className="">{textNoIssues}</div>;
};

export default NotificationModal;
