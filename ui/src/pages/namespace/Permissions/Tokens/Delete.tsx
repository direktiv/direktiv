import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { useDeleteToken } from "~/api/enterprise/tokens/mutate/delete";
import { useTranslation } from "react-i18next";

type DeleteProps = {
  tokenName: string;
  close: () => void;
};

const Delete = ({ tokenName, close }: DeleteProps) => {
  const { t } = useTranslation();
  const { mutate: deleteToken, isPending } = useDeleteToken({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("pages.permissions.tokens.delete.title")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">{t("pages.permissions.tokens.delete.msg")}</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("pages.permissions.tokens.delete.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          onClick={() => {
            deleteToken(tokenName);
          }}
          variant="destructive"
          loading={isPending}
        >
          {!isPending && <Trash />}
          {t("pages.permissions.tokens.delete.deleteBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default Delete;
