import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import {
  RevisionSchemaType,
  revisionConditionNames,
} from "~/api/services/schema/revisions";
import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { FC } from "react";
import { SizeSchema } from "~/api/services/schema";
import { StatusBadge } from "../components/StatusBadge";
import moment from "moment";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const ServicesTableRow: FC<{
  service: string;
  revision: RevisionSchemaType;
  workflow?: string;
  version?: string;
  // not passing a function will disable the delete button
  setDeleteRevision?: (service: RevisionSchemaType | undefined) => void;
}> = ({ revision, service, workflow, version, setDeleteRevision }) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { t } = useTranslation();

  // quick and dirty fix because backend will return string or number
  // depending on the endpoint consumed
  const createdNumber =
    typeof revision.created === "number"
      ? revision.created
      : Number(revision.created);

  const createdAtDate = moment.unix(createdNumber);
  const createdAt = useUpdatedAt(createdAtDate);

  if (!namespace) return null;

  const sizeParse = SizeSchema.safeParse(revision.size);
  const sizeLabel = sizeParse.success
    ? t(`pages.services.revision.create.sizeValues.${sizeParse.data}`)
    : "";

  return (
    <TooltipProvider>
      <TableRow
        onClick={() => {
          if (workflow && version) {
            return navigate(
              pages.explorer.createHref({
                namespace,
                path: workflow,
                serviceName: service,
                serviceVersion: version,
                serviceRevision: revision.rev,
                subpage: "workflow-services",
              })
            );
          }
          return navigate(
            pages.services.createHref({
              namespace,
              service,
              revision: revision.revision,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell>
          <div className="flex flex-col gap-3">
            {revision.name}
            <div className="flex gap-3">
              {revisionConditionNames.map((condition) => {
                const res = (revision.conditions ?? []).find(
                  (c) => c.name === condition
                );
                return (
                  <StatusBadge
                    key={condition}
                    status={res?.status ?? "Unknown"}
                    title={res?.reason ?? undefined}
                    message={res?.message ?? undefined}
                    className="w-fit"
                  >
                    {condition}
                  </StatusBadge>
                );
              })}
            </div>
          </div>
        </TableCell>
        <TableCell>{revision.image}</TableCell>
        <TableCell>{revision.minScale}</TableCell>
        <TableCell>{sizeLabel}</TableCell>
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
              {t("pages.services.revision.list.relativeTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{createdAtDate.format()}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
          {setDeleteRevision && (
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
                  onClick={(e) => {
                    e.stopPropagation();
                    setDeleteRevision(revision);
                  }}
                >
                  <DropdownMenuItem>
                    <Trash className="mr-2 h-4 w-4" />
                    {t("pages.services.revision.list.contextMenu.delete")}
                  </DropdownMenuItem>
                </DialogTrigger>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default ServicesTableRow;
