import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const EventsBreadcrumb = () => {
  const namespace = useNamespace();
  const { isEventsPage } = pages.events.useParams();
  const { icon: Icon } = pages.events;
  const { t } = useTranslation();

  if (!isEventsPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.events.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.events")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default EventsBreadcrumb;
