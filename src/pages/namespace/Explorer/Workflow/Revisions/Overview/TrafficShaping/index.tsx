import { FC, useEffect, useState } from "react";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Network } from "lucide-react";
import RevisionSelector from "./RevisionSelector";
import { Slider } from "~/design/Slider";
import { pages } from "~/util/router/pages";
import { useNodeRevisions } from "~/api/tree/query/revisions";
import { useNodeTags } from "~/api/tree/query/tags";
import { useRouter } from "~/api/tree/query/router";
import { useSetRouter } from "~/api/tree/mutate/setRouter";

// give tags and revisions a shared loading and fetched state
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
  const isLoading = isLoadingTags && isLoadingRevisions;
  const isFetched = isFetchedTags && isFetchedRevisions;
  return { tags, revisions, isLoading, isFetched };
};

const TrafficShaping: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.explorer.useParams();
  const { mutate: setRouter, isLoading: isLoadingMutation } = useSetRouter();
  const { data: router } = useRouter({ path });
  const { tags, revisions, isLoading, isFetched } = useRevisionsAndTags({
    path,
  });

  // local state
  const [routeA, setRouteA] = useState("");
  const [routeB, setRouteB] = useState("");
  const [weight, setWeight] = useState(0);

  // server state
  const routeAServer = router?.routes?.[0]?.ref;
  const routeBServer = router?.routes?.[1]?.ref;
  const weightServer = router?.routes?.[0]?.weight;

  // synch component state with server state
  useEffect(() => {
    if (weightServer) setWeight(weightServer);
    if (routeAServer) setRouteA(routeAServer);
    if (routeBServer) setRouteB(routeBServer);
  }, [weightServer, routeAServer, routeBServer]);

  const isEnabled = routeA && routeB && routeA !== routeB;

  const saveButtonDisabled =
    isLoadingMutation ||
    routeA.length === 0 ||
    routeB.length === 0 ||
    routeA === routeB;

  if (!path) return null;

  // wait for server data to avoid layout shifts
  if (!isFetched) return null;

  return (
    <section className="flex flex-col gap-4">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Network className="h-5" />
        {t(
          "pages.explorer.tree.workflow.revisions.overview.trafficShaping.title"
        )}
      </h3>
      <Card
        className="flex flex-col gap-y-6 p-4"
        data-testid="traffic-shaping-container"
      >
        <div className="text-sm font-medium">
          {t(
            "pages.explorer.tree.workflow.revisions.overview.trafficShaping.description"
          )}
        </div>
        <div className="flex flex-col gap-3 sm:flex-row">
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            defaultValue={routeAServer}
            onSelect={setRouteA}
            data-testid="route-a-selector"
          />
          <RevisionSelector
            className="flex w-full"
            tags={tags?.results ?? []}
            defaultValue={routeBServer}
            revisions={revisions?.results ?? []}
            isLoading={isLoading}
            onSelect={setRouteB}
            data-testid="route-b-selector"
          />
          <div className="flex w-full">
            <Slider
              step={1}
              min={0}
              max={100}
              value={[weight]}
              onValueChange={(e) => {
                setWeight(e[0] ?? 0);
              }}
              data-testid="traffic-shaping-slider"
            />
          </div>
          <div>
            <Button
              block
              disabled={saveButtonDisabled}
              data-testid="traffic-shaping-save-btn"
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
              {t(
                "pages.explorer.tree.workflow.revisions.overview.trafficShaping.saveBtn"
              )}
            </Button>
          </div>
        </div>
        <div className="text-sm" data-testid="traffic-shaping-note">
          {isEnabled ? (
            <Trans
              i18nKey="pages.explorer.tree.workflow.revisions.overview.trafficShaping.setup"
              values={{
                aName: routeA,
                bName: routeB,
                aWeight: weight,
                bWeight: 100 - weight,
              }}
            />
          ) : (
            t(
              "pages.explorer.tree.workflow.revisions.overview.trafficShaping.hint"
            )
          )}
        </div>
      </Card>
    </section>
  );
};

export default TrafficShaping;
