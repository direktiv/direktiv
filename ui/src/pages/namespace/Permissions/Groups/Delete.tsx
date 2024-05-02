import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { GroupSchemaType } from "~/api/enterprise/groups/schema";
import { Trash } from "lucide-react";
import { useDeleteGroup } from "~/api/enterprise/groups/mutation/delete";

const Delete = ({
  group,
  close,
}: {
  group: GroupSchemaType;
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
          <Trash /> {t("pages.permissions.groups.delete.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.permissions.tokens.delete.msg"
          values={{ name: group.group }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.groups.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            deleteGroup(group);
          }}
          variant="destructive"
          loading={isPending}
        >
          {!isPending && <Trash />}
          {t("pages.permissions.groups.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
