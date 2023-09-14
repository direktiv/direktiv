import { Layers, RefreshCcw, RotateCcw } from "lucide-react";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";
import { Trans, useTranslation } from "react-i18next";

import { Card } from "~/design/Card";
import Delete from "~/pages/namespace/Services/List/Delete";
import { Dialog } from "@radix-ui/react-dialog";
import { DialogContent } from "~/design/Dialog";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";

const ServicesList = ({ workflow }: { workflow: string }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();

  const {
    data,
    isSuccess: servicesIsSuccess,
    isAllowed,
    noPermissionMessage,
  } = useServices({
    workflow,
  });

  const DeleteMenuItem = () => {
    const { t } = useTranslation();
    return (
      <>
        <RefreshCcw className="mr-2 h-4 w-4" />
        {t("pages.explorer.tree.workflow.overview.services.deleteMenuItem")}
      </>
    );
  };

  const { t } = useTranslation();

  return (
    <div className="flex flex-col space-y-10 p-5">
      <section className="flex flex-col gap-4">
        <ServicesStreamingSubscriber workflow={workflow} />
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <h3 className="flex items-center gap-x-2 font-bold">
            <Layers className="h-5" />
            {t("pages.explorer.tree.workflow.overview.services.header")}
          </h3>

          <Card className="flex flex-col gap-y-6">
            <ServicesTable
              items={data}
              isSuccess={servicesIsSuccess}
              setDeleteService={setDeleteService}
              deleteMenuItem={<DeleteMenuItem />}
              workflow={workflow}
              isAllowed={isAllowed}
              noPermissionMessage={noPermissionMessage}
            />
          </Card>

          <DialogContent>
            {deleteService && (
              <Delete
                icon={RotateCcw}
                header={t(
                  "pages.explorer.tree.workflow.overview.services.delete.title"
                )}
                message={
                  <Trans
                    i18nKey="pages.explorer.tree.workflow.overview.services.delete.message"
                    values={{ name: deleteService.info.name }}
                  />
                }
                service={deleteService.info.name}
                workflow={workflow}
                version={deleteService.info.revision}
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
