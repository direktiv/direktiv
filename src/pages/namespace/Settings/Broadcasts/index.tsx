import { Card } from "~/design/Card";
import { FC } from "react";
import ItemRow from "../components/ItemRow";
import { Radio } from "lucide-react";
import { Table } from "~/design/Table";
import { useBroadcasts } from "~/api/broadcasts/query/useBroadcasts";
import { useTranslation } from "react-i18next";

const Broadcasts: FC = () => {
  const { t } = useTranslation();

  const { data } = useBroadcasts();

  if (!data?.broadcast) return null;

  const broadcasts = Object.entries(data.broadcast).map(([key, value]) => ({
    name: key,
    value,
  }));

  return (
    <>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Radio className="h-5" />
          {t("pages.settings.broadcasts.list.title")}
        </h3>
      </div>

      <Card>
        <Table>
          {/* TODO: Table layout with toggles, this is temporary */}
          {broadcasts.map((item, i) => (
            <ItemRow key={i} item={item} onDelete={() => "not implemented"} />
          ))}
        </Table>
      </Card>
    </>
  );
};

export default Broadcasts;
