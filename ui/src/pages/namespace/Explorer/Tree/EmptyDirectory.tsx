import { FC } from "react";
import { FolderOpen } from "lucide-react";
import { NewFileDialog } from "./NewFile";
import { NoResult as NoResultContainer } from "~/design/Table";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

const EmptyDirectoryButton = () => {
  const { _splat: path } = useParams({ strict: false });

  return (
    <div className="grid gap-5">
      <NewFileDialog path={path} />
    </div>
  );
};

const EmptyDirectory: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5">
      <NoResultContainer icon={FolderOpen} button={<EmptyDirectoryButton />}>
        {t("pages.explorer.tree.list.empty.title")}
      </NoResultContainer>
    </div>
  );
};

export default EmptyDirectory;
