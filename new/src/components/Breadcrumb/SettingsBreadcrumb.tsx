import { Breadcrumb as BreadcrumbLink } from "~/design/Breadcrumbs";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useTranslation } from "react-i18next";

const SettingsBreadcrumb = () => {
  const namespace = useNamespace();
  const { isSettingsPage } = pages.settings.useParams();
  const { icon: Icon } = pages.settings;
  const { t } = useTranslation();

  if (!isSettingsPage) return null;
  if (!namespace) return null;

  return (
    <>
      <BreadcrumbLink>
        <Link
          to={pages.settings.createHref({
            namespace,
          })}
        >
          <Icon aria-hidden="true" />
          {t("components.breadcrumb.settings")}
        </Link>
      </BreadcrumbLink>
    </>
  );
};

export default SettingsBreadcrumb;
