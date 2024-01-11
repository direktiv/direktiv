import { DropdownMenuItem } from "~/design/Dropdown";
import { LogOut } from "lucide-react";
import enterpriseConfig from "~/config/enterprise";
import { useTranslation } from "react-i18next";

const EnterpriseLogoutButton = () => {
  const { t } = useTranslation();

  return (
    <DropdownMenuItem className="cursor-pointer" asChild>
      <a href={enterpriseConfig.logoutPath} className="flex items-center">
        <LogOut className="mr-2 h-4 w-4" />
        <span>{t("components.userMenu.logout")}</span>
      </a>
    </DropdownMenuItem>
  );
};

export default EnterpriseLogoutButton;
