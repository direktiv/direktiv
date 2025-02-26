import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { RoleSchemaType } from "~/api/enterprise/roles/schema";
import { Trash } from "lucide-react";
import { useDeleteGroup } from "~/api/enterprise/roles/mutation/delete";

const Delete = ({
  group,
  close,
}: {
  group: RoleSchemaType;
  close: () => void;
}) => {
  const { t } = useTranslation();
  const { mutate: deleteGroup, isPending } = useDeleteGroup({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("pages.permissions.roles.delete.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.permissions.roles.delete.msg"
          values={{ name: group.name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.roles.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            deleteGroup(group.name);
          }}
          variant="destructive"
          loading={isPending}
        >
          {!isPending && <Trash />}
          {t("pages.permissions.roles.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
