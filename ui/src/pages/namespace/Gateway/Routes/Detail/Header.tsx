import { FileSymlink, Square } from "lucide-react";

import { AllowAnonymous } from "../components/Anonymous";
import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { Link } from "react-router-dom";
import MessagesOverlay from "../components/MessagesOverlay";
import { Methods } from "../components/Methods";
import Plugins from "../components/Plugins";
import PublicPathInput from "../components/PublicPath";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useRoute } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";

const Header = () => {
  const namespace = useNamespace();
  const { routePath } = pages.gateway.useParams();
  const { data: route } = useRoute({
    routePath: routePath ?? "",
    enabled: !!routePath,
  });

  const { t } = useTranslation();

  if (!route) return null;
  if (!namespace) return null;

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Square className="h-5" /> {route.file_path}
          </h3>
          <div className="flex gap-1">
            <MessagesOverlay messages={route.errors} variant="error">
              {(errorCount) => (
                <Badge variant="destructive">
                  {t("pages.gateway.routes.row.error.count", {
                    count: errorCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
            <MessagesOverlay messages={route.warnings} variant="warning">
              {(warningCount) => (
                <Badge variant="secondary">
                  {t("pages.gateway.routes.row.warnings.count", {
                    count: warningCount,
                  })}
                </Badge>
              )}
            </MessagesOverlay>
          </div>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.gateway.routes.columns.methods")}
          </div>
          <Methods methods={route.methods} />
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.gateway.routes.columns.plugins")}
          </div>
          <Plugins plugins={route.plugins} />
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.gateway.routes.columns.anonymous")}
          </div>
          <AllowAnonymous allow={route.allow_anonymous} />
        </div>
        <div className="grow text-sm">
          {route.path && <PublicPathInput path={route.path} />}
        </div>
        <div className="flex gap-5">
          <Button asChild isAnchor variant="primary" className="max-md:w-full">
            <Link
              to={pages.explorer.createHref({
                namespace,
                subpage: "endpoint",
                path: route.file_path,
              })}
            >
              <FileSymlink />
              {t("pages.gateway.routes.detail.editRoute")}
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Header;
