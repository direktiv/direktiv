import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Settings } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const SettingsBreadcrumb = () => {
  const namespace = useNamespace();

  const isSettingsPage = useMatch({
    from: "/n/$namespace/settings",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isSettingsPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/settings" params={{ namespace }}>
          <Settings aria-hidden="true" />
          {t("components.breadcrumb.settings")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default SettingsBreadcrumb;
