import { DeleteMenuItem, ServiceDelete } from "../components/ServiceDelete";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";

import { Card } from "~/design/Card";
import { Dialog } from "@radix-ui/react-dialog";
import { DialogContent } from "~/design/Dialog";
import { Layers } from "lucide-react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "~/pages/namespace/Services/List/Table";
import { useState } from "react";
import { useTranslation } from "react-i18next";

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
              <ServiceDelete
                service={deleteService}
                workflow={workflow}
                onClose={() => setDialogOpen(false)}
              />
            )}
          </DialogContent>
        </Dialog>
      </section>
    </div>
  );
};

export default ServicesList;
