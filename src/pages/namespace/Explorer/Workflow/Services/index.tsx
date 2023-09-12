import { FC, useState } from "react";
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
import { pages } from "~/util/router/pages";

const WorkflowServicesPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { t } = useTranslation();

  const { data: servicesData, isSuccess: servicesIsSuccess } = useServices({
    workflow: path,
  });

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();

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
    <div className="flex flex-col space-y-10 p-5">
      <Card className="md:col-span-2">
        <ServicesStreamingSubscriber workflow={path} />
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
            <Layers className="h-5" />
            <h3 className="grow">
              {t("pages.explorer.tree.workflow.overview.services.header")}
            </h3>
          </div>

          <ServicesTable
            items={servicesData}
            isSuccess={servicesIsSuccess}
            setDeleteService={setDeleteService}
            deleteMenuItem={<DeleteMenuItem />}
            workflow={path}
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
                workflow={path}
                version={deleteService.info.revision}
                close={() => {
                  setDialogOpen(false);
                }}
              />
            )}
          </DialogContent>
        </Dialog>
      </Card>
    </div>
  );
};

export default WorkflowServicesPage;
