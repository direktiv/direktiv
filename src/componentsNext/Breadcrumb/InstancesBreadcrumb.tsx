import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const InstancesBreadcrumb = () => {
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
        <a
          href={pages.instances.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.instances")}
        </a>
      </BreadcrumbLink>
      {isInstanceDetailPage && instance ? (
        <BreadcrumbLink>
          <Icon aria-hidden="true" />
          <a>{instance.slice(0.8)}</a>
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

export default InstancesBreadcrumb;
