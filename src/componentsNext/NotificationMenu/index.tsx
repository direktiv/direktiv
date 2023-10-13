import Notification from "~/design/Notification";
import { NotificationNoresults } from "~/design/Notification/NotificationModal";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useTranslation } from "react-i18next";

interface NotificationMenuProps {
  className?: string;
}

const NotificationMenu: React.FC<NotificationMenuProps> = () => {
  const { t } = useTranslation();
  const { data, isLoading } = useNamespaceLinting();
  const showIndicator = !!data?.issues.length;

  return (
    <Notification showIndicator={showIndicator} isLoading={isLoading}>
      {/* loading */}
      <NotificationNoresults>
        {t("components.notificationMenu.isLoading.text")}
      </NotificationNoresults>
      <div className="border-red-500"></div>
      {/* no results */}
      <div className="border-red-900"></div>
      {/* has results */}
      <div className="border-yellow-500"></div>
      Test
    </Notification>
  );
};

export default NotificationMenu;
