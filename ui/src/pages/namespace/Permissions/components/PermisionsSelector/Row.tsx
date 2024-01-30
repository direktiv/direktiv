import { TableCell, TableRow } from "~/design/Table";

import { Checkbox } from "~/design/Checkbox";
import { joinPermissionString } from "./utils";

type PermissionRowProps = {
  resource: string;
  scopes: string[];
  availableScopes: string[];
  selectedPermissions: string[];
  onCheckedChange: (permissionValue: string, isChecked: boolean) => void;
};

export const PermissionRow = ({
  resource,
  scopes,
  availableScopes,
  selectedPermissions,
  onCheckedChange,
}: PermissionRowProps) => (
  <TableRow className="lg:pr-8 xl:pr-12">
    <TableCell className="grow">{resource}</TableCell>
    {availableScopes.map((availableScope) => {
      const permissionString = joinPermissionString(availableScope, resource);
      return (
        <TableCell key={availableScope} className="px-2 text-center">
          {scopes.includes(availableScope) && (
            <Checkbox
              checked={selectedPermissions.includes(permissionString)}
              onCheckedChange={(checked) => {
                if (checked !== "indeterminate") {
                  onCheckedChange(permissionString, checked);
                }
              }}
            />
          )}
        </TableCell>
      );
    })}
  </TableRow>
);
