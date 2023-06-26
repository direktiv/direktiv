import { ComponentProps, FC } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Stats, stat } from "fs";
import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { InstanceSchemaType } from "~/api/instances/schema";
import { pages } from "~/util/router/pages";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];
type InstanceStatus = InstanceSchemaType["status"];

const statusToBadgeVariant = (status: InstanceStatus): BadgeVariant => {
  switch (status) {
    case "complete":
      return "success";
    case "crashed":
    case "failed":
      return "destructive";
    case "pending":
      return undefined;
    default:
      break;
  }
};

const InstanceTableRow: FC<{
  instance: InstanceSchemaType;
  namespace: string;
}> = ({ instance, namespace }) => {
  const [name, revision] = instance.as.split(":");
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const createdAt = useUpdatedAt(instance.createdAt);
  const navigate = useNavigate();

  const isLatestRevision = revision === "latest";

  return (
    <TooltipProvider>
      <TableRow
        key={instance.id}
        onClick={() => {
          navigate(
            pages.instances.createHref({
              namespace,
              instance: instance.id,
            })
          );
        }}
        className="cursor-pointer"
      >
        <TableCell>
          <Tooltip>
            <TooltipTrigger>
              <Link
                onClick={(e) => {
                  e.stopPropagation(); // prevent the onClick on the row from firing
                }}
                to={pages.explorer.createHref({
                  namespace,
                  path: name,
                  subpage: isLatestRevision ? "workflow" : "workflow-revisions",
                  revision: isLatestRevision ? undefined : revision,
                })}
                className="hover:underline"
              >
                {name}
              </Link>
            </TooltipTrigger>
            <TooltipContent>click to open workflow</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell
          className="w-32"
          onClick={(e) => {
            e.stopPropagation();
          }}
        >
          <Badge variant="outline">{instance.id.slice(0, 8)}</Badge>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant="outline">{revision}</Badge>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant={statusToBadgeVariant(instance.status)}>
            {instance.status}
          </Badge>
        </TableCell>
        <TableCell className="w-40">
          <Tooltip>
            <TooltipTrigger>{createdAt} ago</TooltipTrigger>
            <TooltipContent>{instance.createdAt}</TooltipContent>
          </Tooltip>
        </TableCell>
        <TableCell className="w-40">
          <Tooltip>
            <TooltipTrigger>{updatedAt} ago</TooltipTrigger>
            <TooltipContent>{instance.updatedAt}</TooltipContent>
          </Tooltip>
        </TableCell>
      </TableRow>
    </TooltipProvider>
  );
};

export default InstanceTableRow;
