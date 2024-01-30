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
  <TableRow>
    <TableCell className="grow">{resource}</TableCell>
    {availableScopes.map((availableScope) => {
      const permissionString = joinPermissionString(availableScope, resource);
      return (
        <TableCell key={availableScope} className="px-2">
          <div className="flex justify-center">
            {scopes.includes(availableScope) ? (
              <Checkbox
                checked={selectedPermissions.includes(permissionString)}
                onCheckedChange={(checked) => {
                  if (checked !== "indeterminate") {
                    onCheckedChange(permissionString, checked);
                  }
                }}
              />
            ) : (
              <Checkbox disabled={true} />
            )}
          </div>
        </TableCell>
      );
    })}
  </TableRow>
);
