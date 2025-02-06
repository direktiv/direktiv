import {
  PermissionMethodAvailableUi,
  PermissionTopic,
  permissionMethodsAvailableUi,
} from "~/api/enterprise/tokens/schema";
import { RadioGroup, RadioGroupItem } from "~/design/RadioGroup";
import { TableCell, TableRow } from "~/design/Table";

import { z } from "zod";

type PermissionRowProps = {
  topic: PermissionTopic;
  onChange: (newValue: PermissionMethodAvailableUi | undefined) => void;
};

const noPermissionsOptionsValue = "";

export const PermissionRow = ({ topic, onChange }: PermissionRowProps) => (
  <RadioGroup
    defaultValue={noPermissionsOptionsValue}
    onValueChange={(newValue) => {
      const parsedValue = z
        .enum(permissionMethodsAvailableUi)
        .safeParse(newValue);

      if (parsedValue.success) {
        onChange(parsedValue.data);
      } else {
        onChange(undefined);
      }
    }}
    className="table-row"
    asChild
  >
    <TableRow>
      <TableCell className="grow">{topic}</TableCell>
      <TableCell className="text-center">
        <RadioGroupItem value={noPermissionsOptionsValue} />
      </TableCell>
      {permissionMethodsAvailableUi.map((permission) => (
        <TableCell key={permission} className="text-center">
          <RadioGroupItem value={permission} />
        </TableCell>
      ))}
    </TableRow>
  </RadioGroup>
);
