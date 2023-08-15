import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { Layers, MoreVertical, PlusCircle, Trash } from "lucide-react";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateService from "./Create";
import Delete from "./Delete";
import NoResult from "./NoResult";
import { useNamespace } from "~/util/store/namespace";
import { useServices } from "~/api/services/query/get";
import { useTranslation } from "react-i18next";

const ServicesListPage = () => {
  const namespace = useNamespace();
  const { data: serviceList, isSuccess } = useServices();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<string>();
  const [createService, setCreateService] = useState(false);

  const { t } = useTranslation();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteService(undefined);
      setCreateService(false);
    }
  }, [dialogOpen]);

  if (!namespace) return null;

  const showTable = (serviceList?.functions?.length ?? 0) > 0;
  const noResults = isSuccess && serviceList?.functions?.length === 0;

  const allAvailableNames =
    serviceList?.functions.map((service) => service.info.name) ?? [];

  const createNewButton = (
    <DialogTrigger asChild>
      <Button onClick={() => setCreateService(true)} variant="outline">
        <PlusCircle />
        {t("pages.services.list.empty.create")}
      </Button>
    </DialogTrigger>
  );

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Layers className="h-5" />
            {t("pages.services.list.title")}
          </h3>
          {createNewButton}
        </div>
        <Card>
          {showTable && (
            <>
              {serviceList?.functions.map((service) => (
                <h1 className="p-2" key={service.info.name}>
                  {service.info.name}{" "}
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={(e) => e.preventDefault()}
                        icon
                      >
                        <MoreVertical />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent className="w-40">
                      <DialogTrigger
                        className="w-full"
                        data-testid="node-actions-delete"
                        onClick={() => {
                          setDeleteService(service.info.name);
                        }}
                      >
                        <DropdownMenuItem>
                          <Trash className="mr-2 h-4 w-4" />
                          {t("pages.services.list.contextMenu.delete")}
                        </DropdownMenuItem>
                      </DialogTrigger>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </h1>
              ))}
            </>
          )}
          {noResults && <NoResult>{createNewButton}</NoResult>}
        </Card>
        <DialogContent>
          {deleteService && (
            <Delete
              service={deleteService}
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
