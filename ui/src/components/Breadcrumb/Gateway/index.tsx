import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import ConsumerBreadcrumb from "./Consumer";
import { Link } from "react-router-dom";
import RoutesBreadcrumb from "./Routes";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const GatewayBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isGatewayPage } = pages.gateway.useParams();
  const { icon: Icon } = pages.gateway;
  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.gateway.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.gateway")}
        </Link>
      </BreadcrumbLink>
      <RoutesBreadcrumb />
      <ConsumerBreadcrumb />
    </>
  );
};

export default GatewayBreadcrumb;
