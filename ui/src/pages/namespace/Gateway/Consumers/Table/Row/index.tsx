import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import { ConsumerSchemaType } from "~/api/gateway/schema";
import { FC } from "react";
import PasswordInput from "./PasswordInput";

type RowProps = {
  consumer: ConsumerSchemaType;
};

export const Row: FC<RowProps> = ({ consumer }) => (
  <TableRow>
    <TableCell>
      <div className="whitespace-normal break-all">{consumer.username}</div>
    </TableCell>
    <TableCell>
      <PasswordInput password={consumer.password} />
    </TableCell>
    <TableCell>
      <PasswordInput password={consumer.api_key} />
    </TableCell>
    <TableCell>
      <div data-testid="consumer-groups" className="flex flex-wrap gap-1">
        {consumer.groups?.map((group) => (
          <Badge key={group} variant="outline">
            {group}
          </Badge>
        ))}
      </div>
    </TableCell>
    <TableCell>
      <div data-testid="consumer-tags" className="flex flex-wrap gap-1">
        {consumer.tags?.map((tag) => (
          <Badge key={tag} variant="outline">
            {tag}
          </Badge>
        ))}
      </div>
    </TableCell>
  </TableRow>
);
