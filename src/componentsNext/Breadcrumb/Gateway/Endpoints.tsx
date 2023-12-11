import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Workflow } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const EndpointBreadcrumb = () => {
  const namespace = useNamespace();
  const { isGatewayEndpointPage } = pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayEndpointPage) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.gateway.createHref({
            namespace,
          })}
        >
          <Workflow aria-hidden="true" />
          {t("components.breadcrumb.gatewayEndpoint")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EndpointBreadcrumb;
