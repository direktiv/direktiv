import { Dialog, DialogContent } from "~/design/Dialog";

import { Card } from "~/design/Card";
import { Layers } from "lucide-react";
import Rebuild from "~/pages/namespace/Services/List/Rebuild";
import RefreshButton from "~/design/RefreshButton";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useWorkflowServices } from "~/api/services/query/services";

const Services = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [rebuildService, setRebuildService] = useState<ServiceSchemaType>();

  const {
    data: serviceList,
    isSuccess,
    isAllowed,
    isFetching,
    refetch,
    noPermissionMessage,
  } = useWorkflowServices(workflow);

  return (
    <Card className="md:col-span-2">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
          <Layers className="h-5" />
          <h3 className="grow">
            {t("pages.explorer.tree.workflow.overview.services.header")}
          </h3>
          <RefreshButton
            icon
            variant="ghost"
            size="sm"
            disabled={isFetching}
            onClick={() => {
              refetch();
            }}
          />
        </div>
        <ServicesTable
          services={serviceList?.data ?? []}
          isSuccess={isSuccess}
          setRebuildService={setRebuildService}
          isAllowed={isAllowed}
          noPermissionMessage={noPermissionMessage}
        />
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
    </Card>
  );
};

export default Services;
