import { FileSymlink, Layers } from "lucide-react";
import { Link, useParams } from "@tanstack/react-router";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { NoPermissions } from "~/design/Table";
import ServiceEditor from "./ServiceEditor";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useNamespaceAndSystemServices } from "~/api/services/query/services";
import { useTranslation } from "react-i18next";

const ServicePage: FC = () => {
  const { _splat: path } = useParams({ strict: false });

  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: serviceData,
    isFetched: isPermissionCheckFetched,
  } = useFile({ path });

  const { data: servicesList } = useNamespaceAndSystemServices();

  if (!path) return null;
  if (serviceData?.type !== "service") return null;
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
                to="/n/$namespace/services/$service"
                from="/n/$namespace"
                params={{ service: serviceId }}
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
