import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { KeyRound } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const TokensBreadcrumb = () => {
  const namespace = useNamespace();
  const isPermissionsTokenPage = useMatch({
    from: "/n/$namespace/permissions/",
    shouldThrow: false,
  });
  const { t } = useTranslation();

  if (!isPermissionsTokenPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions/tokens" params={{ namespace }}>
          <KeyRound aria-hidden="true" />
          {t("components.breadcrumb.permissionsTokens")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default TokensBreadcrumb;
