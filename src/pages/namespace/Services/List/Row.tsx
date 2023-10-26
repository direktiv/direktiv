import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { FC } from "react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { StatusBadge } from "../components/StatusBadge";
import { TooltipProvider } from "~/design/Tooltip";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

const DefaultDeleteMenuItem = () => {
  const { t } = useTranslation();
  return (
    <>
      <DropdownMenuItem>
        <Trash className="mr-2 h-4 w-4" />
        {t("pages.services.list.contextMenu.delete")}
      </DropdownMenuItem>
    </>
  );
};

const ServicesTableRow: FC<{
  service: ServiceSchemaType;
  setDeleteService: (service: ServiceSchemaType) => void;
  deleteMenuItem?: JSX.Element;
  workflow?: string;
}> = ({
  service,
  setDeleteService,
  deleteMenuItem = <DefaultDeleteMenuItem />,
  workflow,
}) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (!namespace) return null;

  return (
    <TooltipProvider>
      <TableRow
        onClick={() => {
          if (workflow) {
            return navigate(
              pages.explorer.createHref({
                namespace,
                path: workflow,
                subpage: "workflow-services",
                //  TODO: serviceName must be renamed to serviceID, revision must be removed from parameters
                serviceName: service.id,
                serviceRevision: "revision",
              })
            );
          }
          return navigate(
            pages.services.createHref({
              namespace,
              service: service.id,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell>
          <div className="flex flex-col gap-3">
            <div>
              {service.name}{" "}
              <span className="whitespace-pre-wrap break-all text-gray-9 dark:text-gray-dark-9">
                {/* TODO: link to the file */}
                {service.filePath}
              </span>
            </div>
            <div className="flex gap-3">
              {service.error && (
                <StatusBadge
                  status="False"
                  className="w-fit"
                  message={service.error}
                >
                  {t("pages.services.list.tableRow.errorLabel")}
                </StatusBadge>
              )}
              {(service.conditions ?? []).map((condition) => (
                <StatusBadge
                  key={condition.type}
                  status={condition.status}
                  message={condition.message}
                  className="w-fit"
                >
                  {condition.type}
                </StatusBadge>
              ))}
            </div>
          </div>
        </TableCell>
        <TableCell>{service.image ? service.image : service.image}</TableCell>
        <TableCell>{service.scale}</TableCell>
        <TableCell>{service.size ? service.size : "-"}</TableCell>
        <TableCell className="whitespace-normal break-all">
          {service.cmd ? service.cmd : "-"}
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
                onClick={(e) => {
                  e.stopPropagation();
                  setDeleteService(service);
                }}
              >
                <DropdownMenuItem>{deleteMenuItem}</DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default ServicesTableRow;
