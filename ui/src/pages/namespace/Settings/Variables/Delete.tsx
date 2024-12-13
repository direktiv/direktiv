import {
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Trash } from "lucide-react";
import { VarSchemaType } from "~/api/variables/schema";

type DeleteProps = {
  items: VarSchemaType[];
  totalItems: number;
  onConfirm: () => void;
};

const Delete = ({ items, totalItems, onConfirm }: DeleteProps) => {
  const { t } = useTranslation();

  const deleteMessage = () => {
    const isSingleItem = items.length === 1;
    const isAllItems = totalItems > 1 && items.length === totalItems;

    if (isSingleItem) {
      return (
        <Trans
          i18nKey="api.variables.mutate.deleteVariable.singleItemMsg"
          values={{ name: items[0]?.name || "" }}
        />
      );
    } else if (isAllItems) {
      return (
        <Trans i18nKey="api.variables.mutate.deleteVariable.allItemsMsg" />
      );
    } else {
      return (
        <Trans
          i18nKey="api.variables.mutate.deleteVariable.multipleItemsMsg"
          values={{ count: items.length }}
        />
      );
    }
  };

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("components.dialog.header.confirm")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">{deleteMessage()}</div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button
          data-testid="registry-delete-confirm"
          onClick={onConfirm}
          variant="destructive"
        >
          {t("components.button.label.delete")}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

export default Delete;
