import { DialogClose, DialogFooter } from "~/design/Dialog";

import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { DialogHeader } from "../components/Header";
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
      <DialogHeader action="delete" path={path} type={block.type} />
      <div>{t("direktivPage.blockEditor.delete.warning")}</div>
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
