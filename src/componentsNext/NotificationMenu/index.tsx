import { FC, PropsWithChildren } from "react";
import {
  NotificationHasresultsButton,
  NotificationHasresultsText,
  NotificationHasresultsTitle,
  NotificationLoading,
  NotificationNoresults,
} from "~/design/Notification/NotificationModal";

import Button from "~/design/Button";
import { DropdownMenuSeparator } from "~/design/Dropdown";
import { Link } from "react-router-dom";
import Notification from "~/design/Notification";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useTranslation } from "react-i18next";

interface NotificationMenuProps {
  className?: string;
}

const NotificationMenu: React.FC<NotificationMenuProps> = ({ className }) => {
  const { t } = useTranslation();
  const { data, isLoading } = useNamespaceLinting();
  const showIndicator = !!data?.issues.length;
  const textLoading = t("components.notificationMenu.isLoading.text");
  const textHasIssues = t("components.notificationMenu.hasIssues.secrets.text");
  const titleHasIssues = t(
    "components.notificationMenu.hasIssues.secrets.title"
  );
  const buttonHasIssues = t(
    "components.notificationMenu.hasIssues.secrets.button"
  );
  const textNoIssues = t("components.notificationMenu.noIssues.text");

  return (
    <div className={twMergeClsx("self-end text-right", className)}>
      <Notification showIndicator={showIndicator} isLoading={isLoading}>
        {/* loading */}
        {isLoading && <NotificationLoading>{textLoading}</NotificationLoading>}
        {/* has results */}
        {showIndicator && !isLoading && (
          <div className="">
            <div className="">
              <NotificationHasresultsTitle>
                {titleHasIssues}
              </NotificationHasresultsTitle>
              <DropdownMenuSeparator className="w-full"></DropdownMenuSeparator>
              <NotificationHasresultsText>
                {textHasIssues}
              </NotificationHasresultsText>
            </div>
            <div className="flex justify-end">
              <NotificationHasresultsButton>
                {buttonHasIssues}
              </NotificationHasresultsButton>
            </div>
          </div>
        )}
        {/* no results */}
        {!showIndicator && !isLoading && (
          <NotificationNoresults>{textNoIssues}</NotificationNoresults>
        )}
      </Notification>
    </div>
  );
};

export default NotificationMenu;
