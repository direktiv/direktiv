import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Users } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const GroupsBreadcrumb = () => {
  const namespace = useNamespace();
  const { isPermissionsGroupPage } = pages.permissions?.useParams() ?? {};
  const { t } = useTranslation();

  if (!isPermissionsGroupPage) return null;
  if (!namespace) return null;
  if (!pages.permissions) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.permissions.createHref({
            namespace,
            subpage: "groups",
          })}
        >
          <Users aria-hidden="true" />
          {t("components.breadcrumb.permissionsGroups")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GroupsBreadcrumb;
