import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { KeyRound } from "lucide-react";
import { useTranslation } from "react-i18next";

const TokensBreadcrumb = () => {
  const isPermissionsTokenPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isPermissionsTokenPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions/tokens" from="/n/$namespace">
          <KeyRound aria-hidden="true" />
          {t("components.breadcrumb.permissionsTokens")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default TokensBreadcrumb;
