import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  Table,
  TableBody,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateServiceRevision from "./Create";
import { pages } from "~/util/router/pages";
import { useServiceDetails } from "~/api/services/query/details";
import { useTranslation } from "react-i18next";

const ServiceDetailPage = () => {
  const { t } = useTranslation();
  const { service } = pages.services.useParams();
  const { data } = useServiceDetails({
    service: service ?? "",
  });

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteService, setDeleteService] = useState<string>();
  const [createService, setCreateService] = useState(false);

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteService(undefined);
      setCreateService(false);
    }
  }, [dialogOpen]);

  if (!service) return null;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Diamond className="h-5" />
            {service}
          </h3>
          <DialogTrigger asChild>
            <Button onClick={() => setCreateService(true)} variant="outline">
              <PlusCircle />
              {t("pages.services.revision.list.create")}
            </Button>
          </DialogTrigger>
        </div>
        <Card>
          <Table>
            <TableHead>
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableHeaderCell>
                  {t("pages.services.list.tableHeader.name")}
                </TableHeaderCell>
                <TableHeaderCell className="w-48">
                  {t("pages.services.list.tableHeader.image")}
                </TableHeaderCell>
                <TableHeaderCell className="w-16">
                  {t("pages.services.list.tableHeader.scale")}
                </TableHeaderCell>
                <TableHeaderCell className="w-20">
                  {t("pages.services.list.tableHeader.size")}
                </TableHeaderCell>
                <TableHeaderCell className="w-48">
                  {t("pages.services.list.tableHeader.cmd")}
                </TableHeaderCell>
                <TableHeaderCell className="w-16" />
              </TableRow>
            </TableHead>
            <TableBody></TableBody>
          </Table>
          {data?.revisions?.map((revision) => (
            <div key={revision.name}>{revision.name}</div>
          ))}
        </Card>

        <DialogContent>
          {/* {deleteService && (
            <Delete
              service={deleteService}
              close={() => {
                setDialogOpen(false);
              }}
            />
          )} */}
          {createService && (
            <CreateServiceRevision
              service={service}
              close={() => setDialogOpen(false)}
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ServiceDetailPage;
