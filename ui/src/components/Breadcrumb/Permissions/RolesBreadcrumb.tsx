import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Users } from "lucide-react";
import { useTranslation } from "react-i18next";

const RolesBreadcrumb = () => {
  const isPermissionsRolesPage = useMatch({
    from: "/n/$namespace/permissions/roles",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isPermissionsRolesPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions/roles" from="/n/$namespace">
          <Users aria-hidden="true" />
          {t("components.breadcrumb.permissionsRoles")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default RolesBreadcrumb;
