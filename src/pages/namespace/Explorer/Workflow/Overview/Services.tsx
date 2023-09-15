import { DeleteMenuItem, ServiceDelete } from "../components/ServiceDelete";
import { Dialog, DialogContent } from "~/design/Dialog";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";

import { Card } from "~/design/Card";
import { Layers } from "lucide-react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const Services = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();

  const { data, isSuccess, isAllowed, noPermissionMessage } = useServices({
    workflow,
  });

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
          isAllowed={isAllowed}
          noPermissionMessage={noPermissionMessage}
        />

        <DialogContent>
          {deleteService && (
            <ServiceDelete
              workflow={workflow}
              service={deleteService}
              onClose={() => setDialogOpen(false)}
            />
          )}
        </DialogContent>
      </Dialog>
    </Card>
  );
};

export default Services;
