import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Layers, PlusCircle, Trash } from "lucide-react";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";
import { Trans, useTranslation } from "react-i18next";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateService from "./Create";
import Delete from "./Delete";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "./Table";

const ServicesListPage = () => {
  const { t } = useTranslation();
  const {
    data: serviceList,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useServices({});

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();
  const [createService, setCreateService] = useState(false);

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteService(undefined);
      setCreateService(false);
    }
  }, [dialogOpen]);

  const allAvailableNames =
    serviceList?.functions.map((service) => service.info.name) ?? [];

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateService(true)} variant="outline">
        <PlusCircle />
        {t("pages.services.list.create")}
      </Button>
    </DialogTrigger>
  );

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <ServicesStreamingSubscriber />

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Layers className="h-5" />
            {t("pages.services.list.title")}
          </h3>
          {createNewButton}
        </div>
        <Card>
          <ServicesTable
            items={serviceList}
            isSuccess={isSuccess}
            setDeleteService={setDeleteService}
            createNewButton={createNewButton}
            isAllowed={isAllowed}
            noPermissionMessage={noPermissionMessage}
          />
        </Card>
        <DialogContent>
          {deleteService && (
            <Delete
              icon={Trash}
              service={deleteService.info.name}
              header={t("pages.services.list.delete.title")}
              message={
                <Trans
                  i18nKey="pages.services.list.delete.msg"
                  values={{ name: deleteService.info.name }}
                />
              }
              close={() => {
                setDialogOpen(false);
              }}
            />
          )}
          {createService && (
            <CreateService
              close={() => setDialogOpen(false)}
              unallowedNames={allAvailableNames}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ServicesListPage;
