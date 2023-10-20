import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Layers, PlusCircle } from "lucide-react";
import {
  ServicesStreamingSubscriber,
  useServices,
} from "~/api/services/query/getAll";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import RefreshButton from "~/design/RefreshButton";
import { ServiceSchemaType } from "~/api/services/schema/services";
import ServicesTable from "./Table";
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
  } = useServices({});

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<ServiceSchemaType>();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteService(undefined);
    }
  }, [dialogOpen]);

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <ServicesStreamingSubscriber />
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Layers className="h-5" />
            {t("pages.services.list.title")}
          </h3>
          <RefreshButton
            icon
            variant="outline"
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
            setDeleteService={setDeleteService}
            isAllowed={isAllowed}
            noPermissionMessage={noPermissionMessage}
          />
        </Card>
        <DialogContent>
          {/* {deleteService && (
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
          )} */}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ServicesListPage;
