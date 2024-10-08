import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { KeyRound } from "lucide-react";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const TokensBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isPermissionsTokenPage } = pages.permissions?.useParams() ?? {};
  const { t } = useTranslation();

  if (!isPermissionsTokenPage) return null;
  if (!namespace) return null;
  if (!pages.permissions) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.permissions.createHref({
            namespace,
            subpage: "tokens",
          })}
        >
          <KeyRound aria-hidden="true" />
          {t("components.breadcrumb.permissionsTokens")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default TokensBreadcrumb;
