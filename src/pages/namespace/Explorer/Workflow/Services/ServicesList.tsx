import { Card } from "~/design/Card";
import { Dialog } from "@radix-ui/react-dialog";
import { DialogContent } from "~/design/Dialog";
import { Layers } from "lucide-react";
import Rebuild from "~/pages/namespace/Services/List/Rebuild";
import RefreshButton from "~/design/RefreshButton";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useWorkflowServices } from "~/api/services/query/services";

const ServicesList = ({ workflow }: { workflow: string }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [rebuildService, setRebuildService] = useState<ServiceSchemaType>();

  const {
    data: serviceList,
    isSuccess,
    isAllowed,
    isFetching,
    noPermissionMessage,
    refetch,
  } = useWorkflowServices(workflow);

  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-10 p-5">
      <section className="flex flex-col gap-4">
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <div className="flex flex-col gap-4 sm:flex-row">
            <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
              <Layers className="h-5" />
              {t("pages.explorer.tree.workflow.overview.services.header")}
            </h3>
            <RefreshButton
              icon
              variant="outline"
              size="sm"
              disabled={isFetching}
              onClick={() => {
                refetch();
              }}
            />
          </div>
          <Card className="flex flex-col gap-y-6">
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
      </section>
    </div>
  );
};

export default ServicesList;
