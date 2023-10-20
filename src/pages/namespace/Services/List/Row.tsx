import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import {
  ServiceSchemaType,
  serviceConditionNames,
} from "~/api/services/schema/services";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { FC } from "react";
import { SizeSchema } from "~/api/services/schema";
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

  const sizeParsed = SizeSchema.safeParse(service.size);
  const sizeLabel = sizeParsed.success
    ? t(`pages.services.create.sizeValues.${sizeParsed.data}`)
    : "";

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
                {service.filePath}
              </span>
            </div>
            <div className="flex gap-3">
              {/* {serviceConditionNames.map((condition) => {
                const res = service.conditions.find(
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
              })} */}
            </div>
          </div>
        </TableCell>
        <TableCell>{service.image}</TableCell>
        <TableCell>{service.scale}</TableCell>
        <TableCell>{sizeLabel}</TableCell>
        <TableCell className="whitespace-normal break-all">
          {service.cmd}
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
