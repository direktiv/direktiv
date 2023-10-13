import { FC, PropsWithChildren } from "react";

import { Loader2 } from "lucide-react";
import { useTranslation } from "react-i18next";

export const NotificationNoresults: FC<PropsWithChildren> = ({ children }) => (
  <div className="">
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
        {/* this can be children and we can use it like this

        <NotificationNoresults>
          no results
        </NotificationNoresults>
        
        */}
        {textLoading}
      </div>
    );
  }

  if (showIndicator) {
    return <div className="">{textHasIssues}</div>;
  }

  return <div className="">{textNoIssues}</div>;

  // return (
  //   <div>
  //     {isLoading && (
  //       <div className="">
  //         <Loader2 className="h-5 animate-spin" />
  //         {textLoading}
  //       </div>
  //     )}

  //     {showIndicator && !isLoading && <div className="">{textHasIssues}</div>}
  //     {!showIndicator && !isLoading && <div className="">{textNoIssues}</div>}
  //   </div>
  // );
};

export default NotificationModal;
