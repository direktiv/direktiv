import { Box } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const InstancesBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isInstancePage, isInstanceDetailPage, instance } =
    pages.instances.useParams();
  const { icon: Icon } = pages.instances;
  const { t } = useTranslation();

  if (!isInstancePage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.instances.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.instances")}
        </Link>
      </BreadcrumbLink>
      {isInstanceDetailPage && instance ? (
        <BreadcrumbLink>
          <Box aria-hidden="true" />
          <Link
            to={pages.instances.createHref({
              namespace,
              instance,
            })}
          >
            {instance.slice(0, 8)}
          </Link>
          <CopyButton
            value={instance}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          />
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

export default InstancesBreadcrumb;
