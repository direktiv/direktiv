import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { FileCheck } from "lucide-react";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const PolicyBreadcrumb = () => {
  const namespace = useNamespace();
  const { isPermissionsPolicyPage } = pages.permissions?.useParams() ?? {};
  const { t } = useTranslation();

  if (!isPermissionsPolicyPage) return null;
  if (!namespace) return null;
  if (!pages.permissions) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.permissions.createHref({
            namespace,
          })}
        >
          <FileCheck aria-hidden="true" />
          {t("components.breadcrumb.permissionsPolicy")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default PolicyBreadcrumb;
