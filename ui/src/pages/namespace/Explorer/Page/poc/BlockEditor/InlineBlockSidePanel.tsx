import { BlockEditFormProps } from ".";
import { CardType } from "../schema/blocks/card";
import { ColumnsType } from "../schema/blocks/columns";
import { Header } from "./components/Header";
import { useTranslation } from "react-i18next";

type QueryProviderEditFormProps = Omit<
  BlockEditFormProps<CardType | ColumnsType>,
  "onSubmit" | "onCancel"
>;

export const InlineBlockSidePanel = ({
  action,
  block,
  path,
}: QueryProviderEditFormProps) => {
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
