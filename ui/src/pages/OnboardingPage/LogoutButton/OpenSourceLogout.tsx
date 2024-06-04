import Button from "~/design/Button";
import { LogOut } from "lucide-react";
import { useApiActions } from "~/util/store/apiKey";
import { useTranslation } from "react-i18next";

const OpenSourceLogoutButton = () => {
  const { t } = useTranslation();
  const { setApiKey: storeApiKey } = useApiActions();

  const logout = () => {
    storeApiKey(null);
  };

  return (
    <Button variant="outline" onClick={logout}>
      <LogOut />
      {t("pages.onboarding.logout")}
    </Button>
  );
};

export default OpenSourceLogoutButton;
