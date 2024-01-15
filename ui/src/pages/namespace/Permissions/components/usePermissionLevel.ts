import Badge from "~/design/Badge";
import { ComponentProps } from "react";

type BadgeVariant = ComponentProps<typeof Badge>["variant"];

const usePermissionLevel = (
  appliedPermissions: string[],
  availablePermissions: string[]
): {
  hasAllPermissions: boolean;
  badgeVariant: BadgeVariant;
} => {
  const noPermissions = appliedPermissions.length === 0;
  const hasAllPermissions =
    !noPermissions &&
    availablePermissions?.every((permission) =>
      appliedPermissions.includes(permission)
    );

  let badgeVariant: BadgeVariant = "secondary";
  if (noPermissions) {
    badgeVariant = "destructive";
  }
  if (hasAllPermissions) {
    badgeVariant = "success";
  }

  return {
    hasAllPermissions,
    badgeVariant,
  };
};

export default usePermissionLevel;
