import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Settings } from "lucide-react";
import { useTranslation } from "react-i18next";

const SettingsBreadcrumb = () => {
  const isSettingsPage = useMatch({
    from: "/n/$namespace/settings",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isSettingsPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/settings" from="/n/$namespace">
          <Settings aria-hidden="true" />
          {t("components.breadcrumb.settings")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default SettingsBreadcrumb;
