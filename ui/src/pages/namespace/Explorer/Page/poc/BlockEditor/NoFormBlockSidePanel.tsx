import { BlockEditFormProps } from ".";
import { Header } from "./components/Header";
import { NoFormBlockType } from "../schema/blocks";
import { useTranslation } from "react-i18next";

type NoFormBlockProps = Omit<
  BlockEditFormProps<NoFormBlockType>,
  "onSubmit" | "onCancel" | "variables"
>;

export const NoFormBlockSidePanel = ({
  action,
  block,
  path,
}: NoFormBlockProps) => {
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
