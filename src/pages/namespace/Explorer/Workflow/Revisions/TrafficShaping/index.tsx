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

const useRevisionsAndTags = ({ path }: { path?: string }) => {
  const {
    data: tags,
    isLoading: isLoadingTags,
    isFetched: isFetchedTags,
  } = useNodeTags({ path });
  const {
    data: revisions,
    isLoading: isLoadingRevisions,
    isFetched: isFetchedRevisions,
  } = useNodeRevisions({
    path,
  });
  const isLoading = isLoadingTags || isLoadingRevisions;
  const isFetched = isFetchedTags && isFetchedRevisions;
  return { tags, revisions, isLoading, isFetched };
};

const TrafficShaping: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.explorer.useParams();
  const { data: router } = useRouter({ path });
  const { tags, revisions, isLoading, isFetched } = useRevisionsAndTags({
    path,
  });

  // component state
  const [routeA, setRouteA] = useState("");
  const [routeB, setRouteB] = useState("");
  const [weight, setWeight] = useState(0);

  // server state
  const { mutate: setRouter, isLoading: isLoadingMutation } = useSetRouter();
  const routeAServer = router?.routes?.[0]?.ref;
  const routeBServer = router?.routes?.[0]?.ref;
  const weightServer = router?.routes?.[0]?.weight;

  // synch component state with server state
  useEffect(() => {
    if (weightServer) setWeight(weightServer);
    if (routeAServer) setRouteA(routeAServer);
    if (routeBServer) setRouteB(routeBServer);
  }, [weightServer, routeAServer, routeBServer]);

  if (!isFetched) return null; // wait for data to to avoid layout shift
  if (!path) return null;

  return (
    <>
      <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
        <Network />
        {t("pages.explorer.tree.workflow.revisions.trafficShaping.title")}
      </h3>
      <Card className="p-4">
        <div className="text-gray-12 dark:text-gray-dark-12">
          <h1>No traffic distribution for this workflow is configured. </h1>
          You can choose two revisions and set a weight between 0 and 100 to
          distribute the traffic between them.
        </div>
        <Separator className="my-4" />
        <div className="flex flex-col gap-3 sm:flex-row">
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            onSelect={setRouteA}
          />
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            onSelect={setRouteB}
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
              disabled={
                isLoadingMutation || routeA.length === 0 || routeB.length === 0
              }
              onClick={() => {
                setRouter({
                  path,
                  routeA: {
                    ref: routeA,
                    weight,
                  },
                  routeB: {
                    ref: routeB,
                    weight: 100 - weight,
                  },
                });
              }}
            >
              Save
            </Button>
          </div>
        </div>
      </Card>
    </>
  );
};

export default TrafficShaping;
