import {
  PermisionSchemaType,
  PermissionMethodAvailableUi,
  PermissionTopic,
} from "~/api/enterprise/tokens/schema";

export const updatePermissions = ({
  permissions,
  value,
  topic,
}: {
  permissions: PermisionSchemaType[];
  topic: PermissionTopic;
  value: PermissionMethodAvailableUi | undefined;
}) => {
  let newPermissions = structuredClone(permissions);
  const permissionAlreadyExists = permissions.some(
    (permission) => permission.topic === topic
  );

  // remove permission
  if (value === undefined) {
    newPermissions = newPermissions.filter(
      (permission) => permission.topic !== topic
    );
  } else {
    // update if permission already exists
    if (permissionAlreadyExists) {
      newPermissions = permissions.map((permission) => {
        if (permission.topic === topic) {
          return { ...permission, method: value };
        }
        return permission;
      });
      // add new permission
    } else {
      newPermissions = [
        ...newPermissions,
        {
          topic,
          method: value,
        },
      ];
    }
  }
  return newPermissions;
};
