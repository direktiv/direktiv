import { BlockPathType } from "../../PageCompiler/Block";
import Button from "~/design/Button";
import { DialogFooter } from "~/design/Dialog";
import { Header } from "../components/Header";
import { useBlock } from "../../PageCompiler/context/pageCompilerContext";
import { useTranslation } from "react-i18next";

type BlockDeleteFormProps = {
  path: BlockPathType;
  onCancel: () => void;
  onSubmit: () => void;
};

export const BlockDeleteForm = ({
  path,
  onCancel,
  onSubmit,
}: BlockDeleteFormProps) => {
  const block = useBlock(path);
  const { t } = useTranslation();

  return (
    <>
      <Header action="delete" path={path} type={block.type} />
      <div className="text-sm">
        {t("direktivPage.blockEditor.delete.warning")}
      </div>
      <DialogFooter>
        <Button variant="ghost" onClick={() => onCancel()}>
          {t("direktivPage.blockEditor.generic.cancelButton")}
        </Button>
        <Button variant="primary" onClick={() => onSubmit()}>
          {t("direktivPage.blockEditor.generic.confirmButton")}
        </Button>
      </DialogFooter>
    </>
  );
};
