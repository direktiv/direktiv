import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";

import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { useBlock } from "../../PageCompiler/context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockDeleteFormProps = {
  path: BlockPathType;
  onSubmit: () => void;
};

export const BlockDeleteForm = ({ path, onSubmit }: BlockDeleteFormProps) => {
  const block = useBlock(path);
  const { t } = useTranslation();

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <div className="flex flex-row justify-between">
            {t("direktivPage.blockEditor.blockForm.title", {
              path: path.join("."),
              action: t("direktivPage.blockEditor.blockForm.action.delete"),
              type: t(`direktivPage.blockEditor.blockForm.type.${block.type}`),
            })}
          </div>
        </DialogTitle>
      </DialogHeader>

      <div className="my-3">{t("direktivPage.blockEditor.delete.warning")}</div>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("direktivPage.blockEditor.generic.cancelButton")}
          </Button>
        </DialogClose>
        <DialogClose asChild>
          <Button variant="primary" onClick={() => onSubmit()}>
            {t("direktivPage.blockEditor.generic.confirmButton")}
          </Button>
        </DialogClose>
      </DialogFooter>
    </>
  );
};
