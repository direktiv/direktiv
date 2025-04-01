import { FC } from "react";
import { NoResult as NoResultContainer } from "~/design/Table";
import { XCircle } from "lucide-react";
import { useTranslation } from "react-i18next";

const NoSearchResult: FC = () => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-col items-center gap-y-5">
      <NoResultContainer icon={XCircle}>
        {t("pages.explorer.tree.list.noSearchResults.title")}
      </NoResultContainer>
    </div>
  );
};

export default NoSearchResult;
