import { FileSymlink, Layers } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { Link } from "react-router-dom";
import { NoPermissions } from "~/design/Table";
import ServiceEditor from "./ServiceEditor";
import { analyzePath } from "~/util/router/utils";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const Header: FC = () => {
  const { path } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: serviceData,
    isFetched: isPermissionCheckFetched,
  } = useNodeContent({ path });

  if (!namespace) return null;
  if (!path) return null;
  if (!serviceData) return null;
  if (!isPermissionCheckFetched) return null;

  if (isAllowed === false)
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <>
      <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5  dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Layers className="h-5" />
            {filename?.relative}
          </h3>
          <Button isAnchor asChild variant="primary">
            <Link
              to={pages.services.createHref({
                namespace,
              })}
            >
              <FileSymlink />
              {t("pages.explorer.service.goToServices")}
            </Link>
          </Button>
        </div>
      </div>
      <ServiceEditor data={serviceData} path={path} />
    </>
  );
};

export default Header;
