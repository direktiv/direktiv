import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import { MoreVertical, Trash } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Button from "~/design/Button";
import { DialogTrigger } from "~/design/Dialog";
import EnvsVariables from "../../components/EnvVariables";
import { FC } from "react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { StatusBadge } from "../../components/StatusBadge";
import { TooltipProvider } from "~/design/Tooltip";
import { useTranslation } from "react-i18next";

const ServicesTableRow: FC<{
  service: ServiceSchemaType;
  setRebuildService: (service: ServiceSchemaType) => void;
}> = ({ service, setRebuildService }) => {
  const { _splat: path } = useParams({ strict: false });
  const navigate = useNavigate();
  const { t } = useTranslation();

  return (
    <TooltipProvider>
      <TableRow
        data-testid="service-row"
        onClick={() => {
          if (service.type === "workflow-service" && path) {
            return navigate({
              to: "/n/$namespace/explorer/workflow/services/$",
              from: "/n/$namespace",
              params: { _splat: path },
            });
          }
          return navigate({
            to: "/n/$namespace/services/$service",
            from: "/n/$namespace",
            params: {
              service: service.id,
            },
          });
        }}
        className="cursor-pointer"
      >
        <TableCell>
          <div className="flex flex-col gap-1">
            <div>
              <span className="whitespace-pre-wrap break-all">
                {service.type === "workflow-service" && path ? (
                  <Link
                    to="/n/$namespace/explorer/workflow/services/$"
                    from="/n/$namespace"
                    params={{ _splat: path }}
                    onClick={(e) => e.stopPropagation()}
                    className="hover:underline"
                  >
                    {service.filePath}
                  </Link>
                ) : (
                  <Link
                    to="/n/$namespace/explorer/service/$"
                    from="/n/$namespace"
                    params={{ _splat: service.filePath }}
                    onClick={(e) => e.stopPropagation()}
                    className="hover:underline"
                  >
                    {service.filePath}
                  </Link>
                )}
              </span>{" "}
              <span className="text-gray-9 dark:text-gray-dark-9">
                {service.name}
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
              <EnvsVariables envs={service.envs} />
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
          {!service.error ? (
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
                    setRebuildService(service);
                  }}
                >
                  <DropdownMenuItem>
                    <Trash className="mr-2 size-4" />
                    {t("pages.services.list.contextMenu.rebuild")}
                  </DropdownMenuItem>
                </DialogTrigger>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : null}
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default ServicesTableRow;
