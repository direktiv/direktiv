import { t } from "i18next";
import { useInstanceId } from "../store/instanceContext";
import { useInstances } from "~/api/instances/query/get";

const maxChildInstancesToShow = 20;

const ChildInstances = () => {
  const instanceId = useInstanceId();
  const { data } = useInstances({
    limit: maxChildInstancesToShow + 1,
    offset: 0,
    filters: {
      TRIGGER: {
        type: "MATCH",
        value: `instance:${instanceId}`,
      },
    },
  });

  const childCount = data?.instances.results.length ?? 0;
  const moreInstances = childCount > maxChildInstancesToShow;

  // need to return an element with a height to avoid layout shifts because of css grid
  if (!data) return <>&nbsp;</>;

  return (
    <div className="text-sm">
      <div className="text-gray-10 dark:text-gray-dark-10">
        {t("pages.instances.list.tableHeader.childInstances.label")}
      </div>
      <span>
        {moreInstances
          ? t(
              "pages.instances.list.tableHeader.childInstances.instanceCountMax",
              {
                count: childCount,
              }
            )
          : t("pages.instances.list.tableHeader.childInstances.instanceCount", {
              count: childCount,
            })}
      </span>
    </div>
  );
};

export default ChildInstances;
