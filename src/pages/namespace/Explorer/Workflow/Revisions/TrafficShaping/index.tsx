import { FC, useState } from "react";

import Button from "../../../../../../design/Button";
import { Card } from "../../../../../../design/Card";
import { Network } from "lucide-react";
import RevisionSelector from "./RevisionSelector";
import { Separator } from "../../../../../../design/Separator";
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
  const [, setA] = useState("");
  const [, setB] = useState("");

  return (
    <>
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Network />
        {t("pages.explorer.tree.workflow.revisions.trafficShaping.title")}
      </h3>
      <Card className="p-4">
        <div className="flex flex-col gap-3 sm:flex-row">
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            onSelect={setA}
          />
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            onSelect={setB}
          />
          <div className="flex w-full">
            <Slider />
          </div>
          <div>
            <Button variant="primary">Save</Button>
          </div>
        </div>
        <Separator className="mt-4" />
      </Card>
    </>
  );
};

export default TrafficShaping;
