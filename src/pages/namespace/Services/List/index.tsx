import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { Layers, MoreVertical, Trash } from "lucide-react";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
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

  const { t } = useTranslation();

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteService(undefined);
    }
  }, [dialogOpen]);

  if (!namespace) return null;

  const showTable = (serviceList?.functions?.length ?? 0) > 0;
  const noResults = isSuccess && serviceList?.functions?.length === 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Layers className="h-5" />
        {t("pages.services.list.title")}
      </h3>
      <Card>
        {showTable && (
          <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
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
            <DialogContent>
              {deleteService && (
                <Delete
                  service={deleteService}
                  close={() => {
                    setDialogOpen(false);
                  }}
                />
              )}
            </DialogContent>
          </Dialog>
        )}
        {noResults && <NoResult />}
      </Card>
    </div>
  );
};

export default ServicesListPage;
