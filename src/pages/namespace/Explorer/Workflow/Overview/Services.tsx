import { Dialog, DialogContent } from "~/design/Dialog";
import { Layers, RefreshCcw, RotateCcw } from "lucide-react";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";
import { Trans, useTranslation } from "react-i18next";

import { Card } from "~/design/Card";
import Delete from "~/pages/namespace/Services/List/Delete";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";

const Services = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();

  const { data, isSuccess } = useServices({
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

  return (
    <Card className="md:col-span-2">
      <ServicesStreamingSubscriber workflow={workflow} />
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
          <Layers className="h-5" />
          <h3 className="grow">
            {t("pages.explorer.tree.workflow.overview.services.header")}
          </h3>
        </div>

        <ServicesTable
          items={data}
          isSuccess={isSuccess}
          setDeleteService={setDeleteService}
          deleteMenuItem={<DeleteMenuItem />}
          workflow={workflow}
        />

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
    </Card>
  );
};

export default Services;
