import { BlockEditFormProps } from ".";
import { Header } from "./components/Header";
import { InlineBlocksType } from "../schema/blocks";
import { useTranslation } from "react-i18next";

type InlineBlockFormProps = Omit<
  BlockEditFormProps<InlineBlocksType>,
  "onSubmit" | "onCancel"
>;

export const InlineBlockSidePanel = ({
  action,
  block,
  path,
}: InlineBlockFormProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col gap-4 px-1">
      <Header action={action} path={path} block={block} />
      <div className="text-gray-10">
        {t("direktivPage.blockEditor.blockForm.noFormDescription")}
      </div>
    </div>
  );
};
