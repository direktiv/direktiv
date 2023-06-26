import { TableCell, TableRow } from "~/design/Table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import Badge from "~/design/Badge";
import { FC } from "react";
import { InstanceSchemaType } from "~/api/instances/schema";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const InstanceTableRow: FC<{
  instance: InstanceSchemaType;
  namespace: string;
}> = ({ instance, namespace }) => {
  const [name, revision] = instance.as.split(":");
  const updatedAt = useUpdatedAt(instance.updatedAt);
  const createdAt = useUpdatedAt(instance.createdAt);

  return (
    <TooltipProvider>
      <TableRow key={instance.id}>
        <TableCell>
          <Link
            to={pages.instances.createHref({
              namespace,
              instance: instance.id,
            })}
          >
            {name}
          </Link>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant="outline">{revision}</Badge>
        </TableCell>
        <TableCell className="w-28">
          <Badge variant="success">{instance.status}</Badge>
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
