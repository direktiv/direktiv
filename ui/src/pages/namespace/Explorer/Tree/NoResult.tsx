import { FC } from "react";
import { FolderOpen } from "lucide-react";
import { NewFileDialog } from "./NewFile";
import { NoResult as NoResultContainer } from "~/design/Table";
import { pages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const EmptyDirectoryButton = () => {
  const { path } = pages.explorer.useParams();

  return (
    <div className="grid gap-5">
      <NewFileDialog path={path} />
    </div>
  );
};

const NoResult: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5">
      <NoResultContainer icon={FolderOpen} button={<EmptyDirectoryButton />}>
        {t("pages.explorer.tree.list.empty.title")}
      </NoResultContainer>
    </div>
  );
};

export default NoResult;
