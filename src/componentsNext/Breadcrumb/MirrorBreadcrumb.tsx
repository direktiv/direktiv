import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
import { GitCompare } from "lucide-react";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const MirrorBreadcrumb = () => {
  const namespace = useNamespace();
  const { isMirrorPage, isActivityDetailPage, activity } =
    pages.mirror.useParams();
  const { icon: Icon } = pages.mirror;
  const { t } = useTranslation();

  if (!isMirrorPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.mirror.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.mirror")}
        </Link>
      </BreadcrumbLink>
      {isActivityDetailPage && activity ? (
        <BreadcrumbLink>
          <GitCompare aria-hidden="true" />
          <Link
            to={pages.mirror.createHref({
              namespace,
              activity,
            })}
          >
            {activity.slice(0, 8)}
          </Link>
          <CopyButton
            value={activity}
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

export default MirrorBreadcrumb;
