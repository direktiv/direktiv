import { AllBlocksType } from "../../schema/blocks";
import { BlockPathType } from "../../PageCompiler/Block";
import { DialogFooter } from "../components/Footer";
import { DialogHeader } from "../components/Header";
import { useTranslation } from "react-i18next";

type BlockDeleteFormProps = {
  action: "delete";
  path: BlockPathType;
  type: AllBlocksType["type"];
  onSubmit: (path: BlockPathType) => void;
};

export const BlockDeleteForm = ({
  action,
  type,
  path,
  onSubmit,
}: BlockDeleteFormProps) => {
  const { t } = useTranslation();

  return (
    <>
      <DialogHeader action={action} path={path} type={type} />
      <div>{t("direktivPage.blockEditor.delete.warning")}</div>
      <DialogFooter onSubmit={() => onSubmit(path)} />
    </>
  );
};
