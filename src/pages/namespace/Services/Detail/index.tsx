import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Diamond, PlusCircle } from "lucide-react";
import {
  NoPermissions,
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";
import {
  ServiceDetailsStreamingSubscriber,
  useServiceDetails,
} from "~/api/services/query/getDetails";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CreateServiceRevision from "./Create";
import Delete from "./Delete";
import { RevisionSchemaType } from "~/api/services/schema/revisions";
import Row from "./Row";
import { pages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const ServiceDetailPage = () => {
  const { t } = useTranslation();
  const { service } = pages.services.useParams();
  const { data, isFetched, isAllowed, noPermissionMessage } = useServiceDetails(
    {
      service: service ?? "",
    }
  );

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteRevision, setDeleteRevision] = useState<RevisionSchemaType>();
  const [createRevision, setCreateRevision] = useState(false);

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteRevision(undefined);
      setCreateRevision(false);
    }
  }, [dialogOpen]);

  if (!service) return null;
  if (!isFetched) return null;

  const showTable = (data?.revisions.length ?? 0) > 0;
  const noResults = isFetched && data?.revisions.length === 0;
  const latestRevision = data?.revisions?.[0];

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <ServiceDetailsStreamingSubscriber service={service} />
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="flex flex-col gap-4 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
            <Diamond className="h-5" />
            {t("pages.services.revision.list.title", { name: service })}
          </h3>
          <DialogTrigger asChild>
            <Button
              onClick={() => setCreateRevision(true)}
              variant="outline"
              // you should not create a new revision when you can't see the table
              disabled={!isAllowed}
            >
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
              {isAllowed ? (
                <>
                  {showTable &&
                    data?.revisions?.map((revision, index) => (
                      <Row
                        revision={revision}
                        service={service}
                        key={revision.name}
                        setDeleteRevision={
                          index !== 0 ? setDeleteRevision : undefined
                        }
                      />
                    ))}
                  {noResults && (
                    <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                      <TableCell colSpan={6}>
                        <NoResult icon={Diamond}>
                          {t("pages.services.revision.list.empty.title")}
                        </NoResult>
                      </TableCell>
                    </TableRow>
                  )}
                </>
              ) : (
                <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                  <TableCell colSpan={6}>
                    <NoPermissions>{noPermissionMessage}</NoPermissions>
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </Card>
        <DialogContent>
          {deleteRevision && (
            <Delete
              service={service}
              revision={deleteRevision}
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
