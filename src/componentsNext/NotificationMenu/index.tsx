import { FC, PropsWithChildren } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  NotificationButton,
  NotificationLoading,
  NotificationText,
  NotificationTitle,
} from "~/design/Notification/NotificationModal";

import Button from "~/design/Button";
import { DropdownMenuSeparator } from "~/design/Dropdown";
import Notification from "~/design/Notification";
import { pages } from "~/util/router/pages";
import { twMergeClsx } from "~/util/helpers";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useTranslation } from "react-i18next";

interface NotificationMenuProps {
  className?: string;
}

// this shall be a function that inserts a Direktiv link in the NotificationButton

const NotificationButtonLink = (linkTo?: string) => {
  let newLink;
  const namespace = useNamespace();
  useTranslation();

  if (!namespace) return ""; // return null ?
  const defaultLink = pages.settings.createHref({
    namespace,
  });

  // delete probably?
  // const navigate = useNavigate();
  // navigate(pages.explorer.createHref({ namespace }));

  // THIS DOES NOT WORK
  /*
  if (!linkTo) {
  newLink = defaultLink;
  } else {
  newLink = pages.{linkTo}.createHref({
  namespace,
  });    
  }

  return newLink;
  */

  return defaultLink;
};

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
      <Notification showIndicator={showIndicator}>
        {/* loading */}
        {isLoading && <NotificationLoading>{textLoading}</NotificationLoading>}
        {/* has results */}
        {showIndicator && !isLoading && (
          <div className="">
            <div className="">
              <NotificationTitle>{titleHasIssues}</NotificationTitle>
              <DropdownMenuSeparator className="w-full"></DropdownMenuSeparator>
              <NotificationText>{textHasIssues}</NotificationText>
            </div>
            <div className="flex justify-end">
              <NotificationButton linkTo="settings">
                {buttonHasIssues}
              </NotificationButton>
            </div>
          </div>
        )}
        {/* no results */}
        {!showIndicator && !isLoading && (
          <NotificationText>{textNoIssues}</NotificationText>
        )}
      </Notification>
    </div>
  );
};

export default NotificationMenu;
