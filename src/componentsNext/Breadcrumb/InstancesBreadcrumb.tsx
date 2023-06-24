import { Box } from "lucide-react";
import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import CopyButton from "~/design/CopyButton";
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
          <Box aria-hidden="true" />
          <a>{instance.slice(0, 8)}</a>
          <CopyButton
            value={instance}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          >
            {(copied) =>
              copied
                ? t(
                    "pages.explorer.tree.workflow.revisions.overview.detail.copied"
                  )
                : t(
                    "pages.explorer.tree.workflow.revisions.overview.detail.copy"
                  )
            }
          </CopyButton>
        </BreadcrumbLink>
      ) : null}
    </>
  );
};

// TODO: check breadcrumb design and margin with > (might also be broken in file browser)
// TODO: link
// TODO: update icon for instances (maybe none=)

export default InstancesBreadcrumb;
