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
import Delete from "./Delete";
import Row from "./Row";
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
  const [deleteRevision, setDeleteRevision] = useState<string>();
  const [createRevision, setCreateRevision] = useState(false);

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRevision(undefined);
      setCreateRevision(false);
    }
  }, [dialogOpen]);

  if (!data) return null;
  if (!service) return null;

  const latestRevision = data.revisions?.[0];

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Diamond className="h-5" />
            {t("pages.services.revision.list.title", { name: service })}
          </h3>
          <DialogTrigger asChild>
            <Button onClick={() => setCreateRevision(true)} variant="outline">
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
                  {t("pages.services.revision.list.tableHeader.name")}
                </TableHeaderCell>
                <TableHeaderCell className="w-48">
                  {t("pages.services.revision.list.tableHeader.image")}
                </TableHeaderCell>
                <TableHeaderCell className="w-16">
                  {t("pages.services.revision.list.tableHeader.scale")}
                </TableHeaderCell>
                <TableHeaderCell className="w-20">
                  {t("pages.services.revision.list.tableHeader.size")}
                </TableHeaderCell>
                <TableHeaderCell className="w-40">
                  {t("pages.services.revision.list.tableHeader.createdAt")}
                </TableHeaderCell>
                <TableHeaderCell className="w-16" />
              </TableRow>
            </TableHead>
            <TableBody>
              {data?.revisions?.map((revision) => (
                <Row
                  revision={revision}
                  service={service}
                  key={revision.name}
                  setDeleteRevision={setDeleteRevision}
                />
              ))}
            </TableBody>
          </Table>
        </Card>
        <DialogContent>
          {deleteRevision && (
            <Delete
              service={deleteRevision}
              close={() => {
                setDialogOpen(false);
              }}
            />
          )}
          {createRevision && (
            <CreateServiceRevision
              service={service}
              close={() => setDialogOpen(false)}
              defaultValues={
                latestRevision
                  ? {
                      image: latestRevision.image,
                      size: latestRevision.size ?? 0,
                      minscale: latestRevision.minScale ?? 0,
                      cmd: "",
                    }
                  : undefined
              }
            />
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
};

export default ServiceDetailPage;
