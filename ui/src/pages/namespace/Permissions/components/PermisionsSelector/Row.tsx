import {
  PermissionMethod,
  PermissionTopic,
  permissionMethodsAvailableUi,
} from "~/api/enterprise/schema";
import { RadioGroup, RadioGroupItem } from "~/design/RadioGroup";
import { TableCell, TableRow } from "~/design/Table";

import { z } from "zod";

type PermissionRowProps = {
  topic: PermissionTopic;
  onChange: (newValue: PermissionMethod | undefined) => void;
  defaultValue?: PermissionMethod;
};

const noPermissionsOptionsValue = "";

export const PermissionRow = ({
  topic,
  onChange,
  defaultValue,
}: PermissionRowProps) => (
  <RadioGroup
    value={defaultValue || noPermissionsOptionsValue}
    onValueChange={(value) => {
      if (value === noPermissionsOptionsValue) {
        onChange(undefined);
        return;
      }
      const parsedValue = z.enum(permissionMethodsAvailableUi).safeParse(value);
      if (parsedValue.success) {
        onChange(parsedValue.data);
        return;
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
