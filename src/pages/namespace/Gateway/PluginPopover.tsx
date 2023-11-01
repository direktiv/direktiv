import { ConditionalWrapper, prettifyJsonString } from "~/util/helpers";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";

import { Card } from "~/design/Card";
import { FC } from "react";
import { GatewaySchemeType } from "~/api/gateway/schema";
import { useTranslation } from "react-i18next";

type RowProps = {
  gateway: GatewaySchemeType;
};
const PluginPopover: FC<RowProps> = ({ gateway }) => {
  const numberOfPlugins = gateway.plugins.length;
  const { t } = useTranslation();

  return (
    <ConditionalWrapper
      condition={numberOfPlugins > 0}
      wrapper={(children) => (
        <HoverCard>
          <HoverCardTrigger>{children}</HoverCardTrigger>
          <HoverCardContent asChild align="end">
            <div className="flex flex-col gap-3">
              {gateway.plugins.map((plugin) => (
                <Card
                  key={plugin.type}
                  className="flex flex-col gap-2 p-5"
                  noShadow
                >
                  <span className="font-bold">{plugin.type}</span>
                  <pre className="whitespace-pre-wrap text-primary-500">
                    {prettifyJsonString(JSON.stringify(plugin.configuration))}
                  </pre>
                </Card>
              ))}
            </div>
          </HoverCardContent>
        </HoverCard>
      )}
    >
      <div>{t("pages.gateway.row.plugins", { count: numberOfPlugins })}</div>
    </ConditionalWrapper>
  );
};

export default PluginPopover;
