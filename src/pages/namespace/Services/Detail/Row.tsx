import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import {
  ServiceRevisionSchemaType,
  serviceRevisionConditionNames,
} from "~/api/services/schema";
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
import { StatusBadge } from "../List/components/StatusBadge";
import moment from "moment";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const ServicesTableRow: FC<{
  service: string;
  revision: ServiceRevisionSchemaType;
  setDeleteRevision: (service: string | undefined) => void;
}> = ({ revision, service, setDeleteRevision }) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const createdAtDate = moment.unix(revision.created);
  const createdAt = useUpdatedAt(createdAtDate);

  if (!namespace) return null;

  const size = revision.size;
  const sizeLabel =
    size === 0 || size === 1 || size === 2
      ? t(`pages.services.revision.create.sizeValues.${size}`)
      : "";

  return (
    <TooltipProvider>
      <TableRow
        onClick={() => {
          navigate(
            pages.services.createHref({
              namespace,
              service,
              revision: revision.name,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell>
          <div className="flex flex-col gap-3">
            {revision.name}
            <div className="flex gap-3">
              {serviceRevisionConditionNames.map((condition) => {
                const res = revision.conditions.find(
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
              {t("pages.instances.list.tableRow.realtiveTime", {
                relativeTime: createdAt,
              })}
            </TooltipTrigger>
            <TooltipContent>{createdAtDate.format()}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell>
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
                  setDeleteRevision(revision.name);
                }}
              >
                <DropdownMenuItem>
                  <Trash className="mr-2 h-4 w-4" />
                  {t("pages.services.revision.list.contextMenu.delete")}
                </DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default ServicesTableRow;
