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
import { useNamespaceServices } from "~/api/services/query/services";
import { useNode } from "~/api/files/query/node";
import { useTranslation } from "react-i18next";

const ServicePage: FC = () => {
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
  } = useNode({ path });

  const { data: servicesList } = useNamespaceServices();

  if (!namespace) return null;
  if (!path) return null;
  if (!serviceData) return null;
  if (!isPermissionCheckFetched) return null;
  if (!servicesList) return null;

  if (isAllowed === false)
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  const serviceId = servicesList.data.find(
    (service) => serviceData.path === service.filePath
  )?.id;

  return (
    <>
      <div className="border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col gap-5 max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Layers className="h-5" />
            {filename?.relative}
          </h3>
          {serviceId && (
            <Button isAnchor asChild variant="primary">
              <Link
                to={pages.services.createHref({
                  namespace,
                  service: serviceId,
                })}
              >
                <FileSymlink />
                {t("pages.explorer.service.goToService")}
              </Link>
            </Button>
          )}
        </div>
      </div>
      <ServiceEditor data={serviceData} />
    </>
  );
};

export default ServicePage;
