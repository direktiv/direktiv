import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Users } from "lucide-react";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const ConsumerBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isGatewayConsumerPage } = pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayConsumerPage) return null;

  return (
    <>
      <BreadcrumbLink data-testid="breadcrumb-consumers">
        <Link
          to={pages.gateway.createHref({
            namespace,
            subpage: "consumers",
          })}
        >
          <Users aria-hidden="true" />
          {t("components.breadcrumb.gatewayConsumers")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default ConsumerBreadcrumb;
