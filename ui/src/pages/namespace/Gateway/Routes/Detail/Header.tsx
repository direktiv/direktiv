import { Link, useParams } from "@tanstack/react-router";
import { Pencil, SquareGanttChartIcon } from "lucide-react";

import { AllowAnonymous } from "../components/Anonymous";
import Badge from "~/design/Badge";
import Button from "~/design/Button";
import MessagesOverlay from "../components/MessagesOverlay";
import { Methods } from "../components/Methods";
import Plugins from "../components/Plugins";
import PublicPathInput from "../components/PublicPath";
import { getMethodsFromOpenApiSpec } from "../utils";
import { useRoute } from "~/api/gateway/query/getRoutes";
import { useTranslation } from "react-i18next";

const Header = () => {
  const { _splat } = useParams({ strict: false });
  const { data: route } = useRoute({
    routePath: _splat ?? "",
    enabled: !!_splat,
  });

  const { t } = useTranslation();

  if (!route) return null;

  return (
    <div
      className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1"
      data-testid="route-details-header"
    >
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <SquareGanttChartIcon className="h-5" />
            {route.file_path}
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
          <Methods methods={getMethodsFromOpenApiSpec(route.spec)} />
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.gateway.routes.columns.plugins")}
          </div>
          <Plugins plugins={route.spec["x-direktiv-config"]?.plugins} />
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.gateway.routes.columns.anonymous")}
          </div>
          <AllowAnonymous
            allow={route.spec["x-direktiv-config"]?.allow_anonymous}
          />
        </div>
        <div className="grow text-sm">
          {route.server_path && <PublicPathInput path={route.server_path} />}
        </div>
        <div className="flex gap-5">
          <Button asChild isAnchor variant="primary" className="max-md:w-full">
            <Link
              to="/n/$namespace/explorer/endpoint/$"
              from="/n/$namespace"
              params={{ _splat }}
            >
              <Pencil />
              {t("pages.gateway.routes.detail.editRoute")}
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
};

export default Header;
