import { Card } from "../../../../../../design/Card";
import { FC } from "react";
import { Network } from "lucide-react";
import RevisionSelector from "./RevisionSelector";
import { Slider } from "../../../../../../design/Slider";
import { pages } from "../../../../../../util/router/pages";
import { useNodeRevisions } from "../../../../../../api/tree/query/revisions";
import { useNodeTags } from "../../../../../../api/tree/query/tags";
import { useTranslation } from "react-i18next";

const TrafficShaping: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.explorer.useParams();
  const { data: revisions, isLoading: revisionsLoading } = useNodeRevisions({
    path,
  });
  const { data: tags, isLoading: tagsLoading } = useNodeTags({ path });

  const isLoading = tagsLoading || revisionsLoading;

  return (
    <>
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Network />
        {t("pages.explorer.tree.workflow.revisions.trafficShaping.title")}
      </h3>
      <Card className="flex gap-x-3 p-4">
        <RevisionSelector
          className="grow"
          tags={tags?.results ?? []}
          revisions={revisions?.results ?? []}
          isLoading={isLoading}
          onSelect={(rev) => console.log(rev)}
        />
        <RevisionSelector
          className="grow"
          tags={tags?.results ?? []}
          revisions={revisions?.results ?? []}
          isLoading={isLoading}
          onSelect={(rev) => console.log(rev)}
        />
        <div className="grow p-4">
          <Slider />
        </div>
      </Card>
    </>
  );
};

export default TrafficShaping;
