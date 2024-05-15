import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const JqPlaygroundBreadcrumb = () => {
  const pages = usePages();
  const namespace = useNamespace();
  const { isJqPlaygroundPage } = pages.jqPlayground.useParams();
  const { icon: Icon } = pages.jqPlayground;
  const { t } = useTranslation();

  if (!isJqPlaygroundPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.jqPlayground.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.jqPlayground")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default JqPlaygroundBreadcrumb;
