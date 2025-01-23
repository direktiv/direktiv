import { BookOpen } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const GatewayInfoBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isGatewayInfoPage } = pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayInfoPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-info">
        <Link to={pages.gateway.createHref({ namespace })}>
          <BookOpen aria-hidden="true" />
          {t("components.breadcrumb.gatewayInfo")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default GatewayInfoBreadcrumb;
