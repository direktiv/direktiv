import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Users } from "lucide-react";
import { useTranslation } from "react-i18next";

const GroupsBreadcrumb = () => {
  const isPermissionsGroupPage = useMatch({
    from: "/n/$namespace/permissions/groups",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isPermissionsGroupPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions/groups" from="/n/$namespace">
          <Users aria-hidden="true" />
          {t("components.breadcrumb.permissionsGroups")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GroupsBreadcrumb;
