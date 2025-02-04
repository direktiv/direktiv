import { Link, useMatch } from "@tanstack/react-router";

import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Users } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const GroupsBreadcrumb = () => {
  const namespace = useNamespace();

  const isPermissionsGroupPage = useMatch({
    from: "/n/$namespace/permissions/groups",
    shouldThrow: false,
  });

  const { t } = useTranslation();

  if (!isPermissionsGroupPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link to="/n/$namespace/permissions/groups" params={{ namespace }}>
          <Users aria-hidden="true" />
          {t("components.breadcrumb.permissionsGroups")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GroupsBreadcrumb;
