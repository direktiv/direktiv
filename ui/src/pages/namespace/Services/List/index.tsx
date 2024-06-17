import { Dialog, DialogContent } from "~/design/Dialog";
import { useEffect, useState } from "react";

import { Card } from "~/design/Card";
import { Layers } from "lucide-react";
import Rebuild from "./Rebuild";
import RefreshButton from "~/design/RefreshButton";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "./Table";
import { useNamespaceAndSystemServices } from "~/api/services/query/services";
import { useTranslation } from "react-i18next";

const ServicesListPage = () => {
  const { t } = useTranslation();
  const {
    data: serviceList,
    isFetching,
    refetch,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useNamespaceAndSystemServices();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [rebuildService, setRebuildService] = useState<ServiceSchemaType>();

  useEffect(() => {
    if (dialogOpen === false) {
      setRebuildService(undefined);
    }
  }, [dialogOpen]);

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Layers className="h-5" />
            {t("pages.services.list.title")}
          </h3>
          <RefreshButton
            icon
            variant="outline"
            aria-label={t("pages.services.list.refetchLabel")}
            disabled={isFetching}
            onClick={() => {
              refetch();
            }}
          />
        </div>
        <Card>
          <ServicesTable
            services={serviceList?.data ?? []}
            isSuccess={isSuccess}
            setRebuildService={setRebuildService}
            isAllowed={isAllowed}
            noPermissionMessage={noPermissionMessage}
          />
        </Card>
        <DialogContent>
          {rebuildService && (
            <Rebuild
              service={rebuildService}
              close={() => {
                setDialogOpen(false);
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ServicesListPage;
