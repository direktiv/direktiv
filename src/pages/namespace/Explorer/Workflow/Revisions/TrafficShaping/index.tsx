import { FC, useEffect, useState } from "react";

import Button from "../../../../../../design/Button";
import { Card } from "../../../../../../design/Card";
import { Network } from "lucide-react";
import RevisionSelector from "./RevisionSelector";
import { Separator } from "../../../../../../design/Separator";
import { Slider } from "../../../../../../design/Slider";
import { pages } from "../../../../../../util/router/pages";
import { useNodeRevisions } from "../../../../../../api/tree/query/revisions";
import { useNodeTags } from "../../../../../../api/tree/query/tags";
import { useRouter } from "~/api/tree/query/router";
import { useSetRouter } from "~/api/tree/mutate/setRouter";
import { useTranslation } from "react-i18next";

const TrafficShaping: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.explorer.useParams();
  const { data: revisions, isLoading: revisionsLoading } = useNodeRevisions({
    path,
  });
  const { data: router } = useRouter({ path });
  const { data: tags, isLoading: tagsLoading } = useNodeTags({ path });
  const { mutate: setRouter, isLoading: isLoadingMutation } = useSetRouter();
  const isLoadingData = tagsLoading || revisionsLoading;

  const [a, setA] = useState("");
  const [b, setB] = useState("");
  const [weight, setWeight] = useState(50);

  const aServer = router?.routes?.[0]?.ref;
  const bServer = router?.routes?.[0]?.ref;
  const weightServer = router?.routes?.[0]?.weight;

  useEffect(() => {
    if (aServer) setA(aServer);
  }, [aServer]);

  useEffect(() => {
    if (bServer) setB(bServer);
  }, [bServer]);

  useEffect(() => {
    if (weightServer) setWeight(weightServer);
  }, [weightServer]);

  if (!path) return null;

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
            isLoading={isLoadingData}
            onSelect={setA}
          />
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoadingData}
            onSelect={setB}
          />
          <div className="flex w-full">
            <Slider
              step={1}
              min={0}
              max={100}
              value={[weight]}
              onValueChange={(e) => {
                setWeight(e?.[0] || 0);
              }}
            />
          </div>
          <div>
            <Button
              variant="primary"
              disabled={isLoadingMutation || a.length === 0 || b.length === 0}
              onClick={() => {
                setRouter({
                  path,
                  routeA: {
                    ref: a,
                    weight,
                  },
                  routeB: {
                    ref: b,
                    weight: 100 - weight,
                  },
                });
              }}
            >
              Save
            </Button>
          </div>
        </div>
        <Separator className="my-4" />
        <div>
          {aServer} - {bServer} - {weightServer ?? "none"}
        </div>
        <div>
          {a} - {b} - {weight ?? "none"}
        </div>
      </Card>
    </>
  );
};

export default TrafficShaping;
