import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { Users } from "lucide-react";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const ConsumerBreadcrumb = () => {
  const namespace = useNamespace();
  const { isGatewayConsumerPage } = pages.gateway.useParams();

  const { t } = useTranslation();

  if (!namespace) return null;
  if (!isGatewayConsumerPage) return null;

  return (
    <>
      <BreadcrumbLink>
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
