import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import Badge from "~/design/Badge";
import { FC } from "react";
import { ServiceSchemaType } from "~/api/services/schema/services";
import { useTranslation } from "react-i18next";

export type EnvsVariablesProps = {
  envs: ServiceSchemaType["envs"];
};

const EnvsVariables: FC<EnvsVariablesProps> = ({ envs }) => {
  const { t } = useTranslation();
  const envsCount = envs?.length ?? 0;

  return envsCount > 0 ? (
    <HoverCard>
      <HoverCardTrigger className="inline-flex">
        <Badge variant="secondary">
          {t("pages.services.list.tableRow.envsLabel", {
            count: envsCount,
          })}
        </Badge>
      </HoverCardTrigger>
      <HoverCardContent className="flex flex-col gap-2 p-3">
        {envs?.map(({ name, value }, i) => (
          <code key={i}>
            {name}={value}
          </code>
        ))}
      </HoverCardContent>
    </HoverCard>
  ) : null;
};

export default EnvsVariables;
