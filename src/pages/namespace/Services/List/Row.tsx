import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { MoreVertical, Trash } from "lucide-react";
import { ServiceSchemaType, conditionNames } from "~/api/services/schema";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import { FC } from "react";
import { StatusBadge } from "./components/StatusBadge";
import { TooltipProvider } from "~/design/Tooltip";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";

const ServicesTableRow: FC<{
  service: ServiceSchemaType;
  setDeleteService: (service: string | undefined) => void;
}> = ({ service, setDeleteService }) => {
  const namespace = useNamespace();
  const navigate = useNavigate();
  const { t } = useTranslation();

  if (!namespace) return null;

  const size = service.info.size;
  const sizeLabel =
    size === 0 || size === 1 || size === 2
      ? t(`pages.services.create.sizeValues.${size}`)
      : "";

  return (
    <TooltipProvider>
      <TableRow
        onClick={() => {
          navigate(
            pages.services.createHref({
              namespace,
              service: service.info.name,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell>{service.info.name}</TableCell>
        <TableCell>{service.info.image}</TableCell>
        <TableCell>{service.info.minScale}</TableCell>
        <TableCell>{sizeLabel}</TableCell>
        <TableCell>{service.info.cmd}</TableCell>
        <TableCell>
          <div className="flex flex-col gap-2">
            {conditionNames.map((condition) => {
              const res = service.conditions.find((c) => c.name === condition);
              if (!res) return null;
              return (
                <StatusBadge
                  key={condition}
                  status={res.status}
                  title={res.reason}
                  message={res.message}
                  className="w-fit"
                >
                  {res.name}
                </StatusBadge>
              );
            })}
          </div>
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
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default ServicesTableRow;
